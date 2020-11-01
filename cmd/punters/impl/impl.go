package impl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cbellee/dapr-bet/cmd/punters/spec"
	"github.com/cbellee/dapr-bet/pkg/helper"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/google/uuid"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// PuntersService implements a dapr service and client
type PuntersService struct {
	client     dapr.Client
	server     common.Service
	components spec.DaprComponents
}

// NewService creates a new instance of the ResultsService
func NewService(serviceName string, servicePort string, components spec.DaprComponents) *PuntersService {
	client, err := dapr.NewClient()
	if err != nil {
		logger.Panicf("Failed to create Dapr client: %s", err)
		return nil
	}

	port := fmt.Sprintf(":%s", servicePort)
	server := daprd.NewService(port)

	service := &PuntersService{
		client,
		server,
		components,
	}

	return service
}

// StartService starts the http server
func (s *PuntersService) StartService() error {
	err := s.server.Start()
	if err != nil {
		return err
	}
	return nil
}

// AddTopicHandler wires up a new topic event handler
func (s *PuntersService) AddTopicHandler(sub *common.Subscription, handler func(context.Context, *common.TopicEvent) error) error {
	err := s.server.AddTopicEventHandler(sub, handler)
	if err != nil {
		logger.Fatalf("Error adding topic event handler: %s", err)
		return err
	}

	return nil
}

// AddServiceHandler wires up a new service invocation handler
func (s *PuntersService) AddServiceHandler(name string, handler func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error {
	err := s.server.AddServiceInvocationHandler(name, handler)
	if err != nil {
		logger.Printf("Error registering service handler: %s", err)
		return err
	}
	return nil
}

// GetPunter gets a punter by email address using a query string, i.e. '/get?email=cbellee@microsoft.com'
func (s *PuntersService) GetPunter(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "getPunter()")

	if in.QueryString["email"] == "" {
		logger.Print("querystring 'email' not provided or nil value, exiting...")
		return
	}

	var email = in.QueryString["email"]
	logger.Printf("QueryString: " + email)

	ctx = context.Background()
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf(err.Error())
	}

	// get state from dapr statestore
	data, err := client.GetStateWithConsistency(ctx, s.components.StateStoreName, email, dapr.StateConsistencyUndefined)
	if err != nil {
		logger.Printf(err.Error())
	}

	if data.Value == nil {
		logger.Printf("no data found for query '%s' in state store '%s'", email, s.components.StateStoreName)
		return nil, nil
	}

	logger.Printf("state data key: %s", data.Key)
	logger.Printf("state data value: %v", string(data.Value))

	out = &common.Content{
		Data:        data.Value,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	return out, nil
}

// AddPunterHandler adds a new punter
func (s *PuntersService) AddPunter(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "addPunter()")

	ctx = context.Background()
	client, err := dapr.NewClient()

	var p spec.Punter
	err = json.Unmarshal(in.Data, &p)
	if err != nil {
		log.Print(err.Error())
	}

	uuid := uuid.New()
	p.ID = uuid.String()

	data, err := client.GetStateWithConsistency(ctx, s.components.StateStoreName, p.Email, dapr.StateConsistencyUndefined)
	if err != nil {
		log.Printf(err.Error())
	}

	if data.Value != nil {
		return nil, errors.New("punter with email '" + p.Email + "' already exists")
	}

	bytArr, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	punter := &dapr.SetStateItem{
		Key:   p.Email,
		Value: bytArr,
		Etag:  p.ID,
		Metadata: map[string]string{
			"created": time.Now().UTC().String(),
		},
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyLastWrite,
			Consistency: dapr.StateConsistencyUndefined,
		},
	}

	if err := client.SaveStateItems(ctx, s.components.StateStoreName, punter); err != nil {
		panic(err)
	}
	log.Printf("new punter '%s' added successfully", p.Email)

	out = &common.Content{
		Data:        []byte(data.Value),
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	return out, nil
}

// balanceHandler credits and debits a punters account balance
func (s *PuntersService) CalculateBalance(ctx context.Context, e *common.TopicEvent) error {
	defer helper.TimeTrack(time.Now(), "balanceHandler()")

	b, err := json.Marshal(e.Data)
	if err != nil {
		log.Fatal(err)
	}

	var bet spec.Bet
	err = json.Unmarshal(b, &bet)
	if err != nil {
		panic(err)
	}

	logger.Printf("amount: %f email: %s horseid: %d raceid: %d place: %d odds: %s, paid: %t", bet.Amount, bet.Email, bet.HorseID, bet.RaceID, bet.Place, bet.Odds, bet.Paid)

	creditBalance(bet, s.components.StateStoreName)
	return nil
}

// debitBalance debits a punters balance
func debitBalance(email string, amount float32, storeName string) error {
	defer helper.TimeTrack(time.Now(), "DebitBalance()")

	ctx := context.Background()
	client, err := dapr.NewClient()
	if err != nil {
		log.Printf(err.Error())
	}

	data, err := client.GetStateWithConsistency(ctx, storeName, email, dapr.StateConsistencyUndefined)
	if err != nil {
		panic(err.Error())
	}

	var p spec.Punter
	err = json.Unmarshal(data.Value, &p)
	if err != nil {
		panic(err.Error())
	}

	if amount <= p.Balance && p.Balance > 0 {
		newBalance := p.Balance - amount
		p.Balance = newBalance
	} else {
		return errors.New("insufficient funds for punter '" + email + "': [" + fmt.Sprintf("%.2f", p.Balance) + "]")
	}

	bytArr, err := json.Marshal(p)
	punter := &dapr.SetStateItem{
		Key: p.Email,
		Metadata: map[string]string{
			"modified": time.Now().UTC().String(),
		},
		Value: bytArr,
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyUndefined,
			Consistency: dapr.StateConsistencyUndefined,
		},
	}

	if err := client.SaveStateItems(ctx, storeName, punter); err != nil {
		panic(err)
	}
	log.Printf("balance debited by $"+fmt.Sprintf("%.2f", amount)+" for punter '%s'", p.Email)
	return nil
}

// creditBalance credits a punters balance
func creditBalance(bet spec.Bet, storeName string) error {
	defer helper.TimeTrack(time.Now(), "creditBalance()")

	ctx := context.Background()
	client, err := dapr.NewClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	data, err := client.GetStateWithConsistency(ctx, storeName, bet.Email, dapr.StateConsistencyUndefined)
	if err != nil {
		panic(err.Error())
	}

	if data.Value == nil {
		log.Printf("Punter not found: '%s'", bet.Email)
		return nil
	}

	var p spec.Punter
	err = json.Unmarshal(data.Value, &p)
	if err != nil {
		panic(err.Error())
	}

	strArr := strings.Split(bet.Odds, "/")
	floatArr := stringToFloatArray(strArr)

	o := floatArr[0] / floatArr[1]
	win := o * bet.Amount
	newBalance := p.Balance + win
	p.Balance = newBalance

	bytArr, err := json.Marshal(p)

	punter := &dapr.SetStateItem{
		Key: p.Email,
		Metadata: map[string]string{
			"modified": time.Now().UTC().String(),
		},
		Value: bytArr,
		Options: &dapr.StateOptions{
			Concurrency: dapr.StateConcurrencyUndefined,
			Consistency: dapr.StateConsistencyUndefined,
		},
	}

	if err := client.SaveStateItems(ctx, storeName, punter); err != nil {
		panic(err)
	}
	log.Printf("balance credited by $"+fmt.Sprintf("%.2f", bet.Amount)+" for punter '%s'", punter.Key)
	return nil
}

// stringToFloatArray converts an array of strings to an array of float32s
func stringToFloatArray(strArr []string) []float32 {
	defer helper.TimeTrack(time.Now(), "stringToFloatArray()")
	var f = []float32{}

	for _, i := range strArr {
		j, err := strconv.ParseFloat(i, 32)
		k := float32(j)
		if err != nil {
			panic(err)
		}
		f = append(f, k)
	}
	return f
}
