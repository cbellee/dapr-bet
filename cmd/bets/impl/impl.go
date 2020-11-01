package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cbellee/dapr-bet/cmd/bets/spec"
	"github.com/cbellee/dapr-bet/pkg/helper"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// BetsService implements a dapr service and client
type BetsService struct {
	client     dapr.Client
	server     common.Service
	dbConfig   spec.DbConfig
	components spec.DaprComponents
}

// NewService creates a new instance of the ResultsService
func NewService(serviceName string, servicePort string, dbConfig spec.DbConfig, components spec.DaprComponents) *BetsService {
	client, err := dapr.NewClient()
	if err != nil {
		logger.Panicf("Failed to create Dapr client: %s", err)
		return nil
	}

	port := fmt.Sprintf(":%s", servicePort)
	server := daprd.NewService(port)

	service := &BetsService{
		client,
		server,
		dbConfig,
		components,
	}

	return service
}

// StartService starts the http server
func (s *BetsService) StartService() error {
	err := s.server.Start()
	if err != nil {
		return err
	}
	return nil
}

// AddTopicHandler wires up a new topic event handler
func (s *BetsService) AddTopicHandler(sub *common.Subscription, handler func(context.Context, *common.TopicEvent) error) error {
	err := s.server.AddTopicEventHandler(sub, handler)
	if err != nil {
		logger.Fatalf("Error adding topic event handler: %s", err)
		return err
	}

	return nil
}

// AddServiceHandler wires up a new service invocation handler
func (s *BetsService) AddServiceHandler(name string, handler func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error {
	err := s.server.AddServiceInvocationHandler(name, handler)
	if err != nil {
		logger.Printf("Error registering service handler: %s", err)
		return err
	}
	return nil
}

// AddBet adds a new bet to the database
func (s *BetsService) AddBet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "addBet()")

	if in == nil {
		logger.Print("getBetHandler input is null, exiting...")
		return
	}

	logger.Printf("/addbet - ContentType:%s, Verb:%s, QueryString:%s", in.ContentType, in.Verb, in.QueryString)

	client, err := dapr.NewClient()
	if err != nil {
		logger.Fatalf("error creating dapr client: %s", err)
	}

	var b spec.Bet
	err = json.Unmarshal(in.Data, &b)
	if err != nil {
		logger.Printf("error un-marshaling JSON: %s", err.Error())
	}

	logger.Printf("%v", b)

	uuid := uuid.New()
	b.ID = uuid.String()
	b.Paid = false

	bytArr, err := json.Marshal(b)
	if err != nil {
		logger.Print(err.Error())
	}

	bi := &dapr.BindingInvocation{
		Name:      s.components.BindingName,
		Data:      bytArr,
		Operation: "create",
	}

	logger.Printf("invoking binding '%s'", s.components.BindingName)
	err = client.InvokeOutputBinding(ctx, bi)
	if err != nil {
		logger.Print(err.Error())
	}

	out = &common.Content{
		Data:        bytArr,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	logger.Printf("new bet with ID: '%s' RaceName: '%s' RaceID: '%d' Amount: '%f' HorseID: '%d' HorseName: '%s' saved successfully for punter '%s'", b.ID, b.RaceName, b.RaceID, b.Amount, b.HorseID, b.HorseName, b.Email)
	return out, nil
}

// GetBet returns a bet from the database
func (s *BetsService) GetBet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "GetBet()")

	var bet spec.Bet
	err = json.Unmarshal(in.Data, &bet)
	if err != nil {
		logger.Print(err.Error())
	}

	cosmosCfg := cosmosapi.Config{
		MasterKey: s.dbConfig.DbKey,
	}

	qry := cosmosapi.Query{
		Query: "SELECT * FROM c WHERE c.id = @betid",
		Params: []cosmosapi.QueryParam{
			{
				Name:  "@betid",
				Value: bet.ID,
			},
		},
	}

	client := cosmosapi.New(s.dbConfig.DbURL, cosmosCfg, nil, nil)
	queryOps := cosmosapi.DefaultQueryDocumentOptions()
	queryOps.IsQuery = true

	var bets []spec.Bet
	queryResponse, err := client.QueryDocuments(context.Background(), s.dbConfig.DbName, s.dbConfig.DbContainer, qry, &bets, queryOps)
	if err != nil {
		err = errors.WithStack(err)
		logger.Print(err)
		return nil, err
	}

	logger.Printf("query returned [%d] documents", queryResponse.Count)

	bytArr, err := json.Marshal(&bets)
	if err != nil {
		logger.Printf(err.Error())
	}

	out = &common.Content{
		Data:        bytArr,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	return out, nil
}

// GetBets returns all bets for the current result from the database
func (s *BetsService) GetBets(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "GetBets()")
	cosmosCfg := cosmosapi.Config{
		MasterKey: s.dbConfig.DbKey,
	}

	client := cosmosapi.New(s.dbConfig.DbURL, cosmosCfg, nil, nil)
	listOps := cosmosapi.ListDocumentsOptions{
		MaxItemCount:        1000,
		AIM:                 "",
		Continuation:        "",
		IfNoneMatch:         "",
		PartitionKeyRangeId: "",
	}

	var bets []spec.Bet
	_, err = client.ListDocuments(context.Background(), s.dbConfig.DbName, s.dbConfig.DbContainer, &listOps, &bets)
	if err != nil {
		err = errors.WithStack(err)
		log.Fatalf("error in listResponse: %s", err.Error())
	}

	bytArr, err := json.Marshal(bets)
	if err != nil {
		log.Println(err)
	}

	out = &common.Content{
		Data:        bytArr,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	return out, nil
}

// GetBetsByEmail returns all bets for a punter from the database
func (s *BetsService) GetBetsByEmail(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "GetBetsByEmail()")

	if in.QueryString["email"] == "" {
		logger.Print("querystring 'email' not provided or nil value, exiting...")
		return
	}

	var email = in.QueryString["email"]
	logger.Printf("QueryString: " + email)

	cosmosCfg := cosmosapi.Config{
		MasterKey: s.dbConfig.DbKey,
	}

	qry := cosmosapi.Query{
		Query: "SELECT * FROM c WHERE c.email = @emailid",
		Params: []cosmosapi.QueryParam{
			{
				Name:  "@emailid",
				Value: email,
			},
		},
	}

	client := cosmosapi.New(s.dbConfig.DbURL, cosmosCfg, nil, nil)
	queryOps := cosmosapi.DefaultQueryDocumentOptions()
	queryOps.EnableCrossPartition = true
	queryOps.IsQuery = true

	var bets []spec.Bet
	queryResponse, err := client.QueryDocuments(context.Background(), s.dbConfig.DbName, s.dbConfig.DbContainer, qry, &bets, queryOps)
	if err != nil {
		err = errors.WithStack(err)
		logger.Print(err)
		return nil, err
	}

	logger.Printf("query returned [%d] documents", queryResponse.Count)

	bytArr, err := json.Marshal(&bets)
	if err != nil {
		logger.Printf(err.Error())
	}

	out = &common.Content{
		Data:        bytArr,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	return out, nil
}

// GetBetResults finds bets from current result entry
func (s *BetsService) GetBetResults(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "GetBetResults()")

	if in == nil {
		return
	}

	cosmosCfg := cosmosapi.Config{
		MasterKey: s.dbConfig.DbKey,
	}

	var b spec.Bet
	err = json.Unmarshal(in.Data, &b)
	if err != nil {
		log.Fatal(err.Error())
	}

	logger.Printf("%v", b)

	qry := cosmosapi.Query{
		Query: "SELECT * FROM c WHERE c.raceid = @raceid AND c.horseid = @horseid",
		Params: []cosmosapi.QueryParam{
			{
				Name:  "@raceid",
				Value: b.RaceID,
			},
			{
				Name:  "@horseid",
				Value: b.HorseID,
			},
		},
	}

	client := cosmosapi.New(s.dbConfig.DbURL, cosmosCfg, nil, nil)
	queryOps := cosmosapi.DefaultQueryDocumentOptions()
	queryOps.EnableCrossPartition = true
	queryOps.IsQuery = true

	var bets []spec.Bet
	queryResponse, err := client.QueryDocuments(context.Background(), s.dbConfig.DbName, s.dbConfig.DbContainer, qry, &bets, queryOps)
	if err != nil {
		err = errors.WithStack(err)
		logger.Print(err)
		return nil, err
	}

	logger.Printf("query returned [%d] documents", queryResponse.Count)

	// only return bet results that aren't empty
	if len(bets) > 0 {
		bytArr, err := json.Marshal(&bets)
		if err != nil {
			log.Printf(err.Error())
		}

		out = &common.Content{
			Data:        bytArr,
			ContentType: in.ContentType,
			DataTypeURL: in.DataTypeURL,
		}

		// send message to 'payments' topic
		for _, b := range bets {
			sendMessageToTopic(s.components.MessageBusName, s.components.TopicName, b)
		}
		return out, nil
	}
	return nil, nil
}

// sendMessageToTopic sends a message a messaging endpoint/topic
func sendMessageToTopic(pubName string, topicName string, bet spec.Bet) error {
	defer helper.TimeTrack(time.Now(), "sendMessageToTopic()")

	ctx := context.Background()
	c, err := dapr.NewClient()
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(bet)
	if err != nil {
		panic(err.Error())
	}

	// send payment information to 'payments' service bus topic
	log.Printf("sending event to topic '%s'", topicName)
	if err := c.PublishEvent(ctx, pubName, topicName, data); err != nil {
		log.Printf(err.Error())
		return err
	}
	return nil
}
