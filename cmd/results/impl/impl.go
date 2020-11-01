package impl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cbellee/dapr-bet/cmd/results/spec"
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

// ResultsService implements a dapr service and client
type ResultsService struct {
	client     dapr.Client
	server     common.Service
	dbConfig   spec.DbConfig
	components spec.DaprComponents
}

// NewService creates a new instance of the ResultsService
func NewService(serviceName string, servicePort string, dbConfig spec.DbConfig, components spec.DaprComponents) *ResultsService {
	client, err := dapr.NewClient()
	if err != nil {
		logger.Panicf("Failed to create Dapr client: %s", err)
		return nil
	}

	port := fmt.Sprintf(":%s", servicePort)
	server := daprd.NewService(port)

	service := &ResultsService{
		client,
		server,
		dbConfig,
		components,
	}

	return service
}

// StartService starts the http server
func (s *ResultsService) StartService() error {
	err := s.server.Start()
	if err != nil {
		return err
	}
	return nil
}

// AddTopicHandler wires up a new topic event handler
func (s *ResultsService) AddTopicHandler(sub *common.Subscription, handler func(context.Context, *common.TopicEvent) error) error {
	err := s.server.AddTopicEventHandler(sub, handler)
	if err != nil {
		logger.Fatalf("Error adding topic event handler: %s", err)
		return err
	}

	return nil
}

// AddServiceHandler wires up a new service invocation handler
func (s *ResultsService) AddServiceHandler(name string, handler func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error {
	err := s.server.AddServiceInvocationHandler(name, handler)
	if err != nil {
		logger.Printf("Error registering service handler: %s", err)
		return err
	}
	return nil
}

// GetResults returns all results in the DB
func (s *ResultsService) GetResults(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "getResultsHandler()")

	cosmosCfg := cosmosapi.Config{
		MasterKey: s.dbConfig.DbKey,
	}

	qry := cosmosapi.Query{
		Query: "SELECT r.raceid, r.racename, h.horseid, h.horsename, h.place, h.odds FROM r JOIN h IN r.runners WHERE  h.place > 0 AND h.place <= 3",
	}

	client := cosmosapi.New(s.dbConfig.DbURL, cosmosCfg, nil, nil)
	queryOps := cosmosapi.DefaultQueryDocumentOptions()
	queryOps.EnableCrossPartition = true
	queryOps.IsQuery = true

	var results []spec.PlacedResult
	queryResponse, err := client.QueryDocuments(context.Background(), s.dbConfig.DbName, s.dbConfig.DbContainer, qry, &results, queryOps)
	if err != nil {
		err = errors.WithStack(err)
		logger.Print(err)
		return nil, err
	}

	logger.Printf("query returned [%d] documents", queryResponse.Count)

	bytArr, err := json.Marshal(&results)
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

// AddResult adds a new result to the database
func (s *ResultsService) AddResult(r spec.Result, bindingName string) error {
	defer helper.TimeTrack(time.Now(), "addResult()")

	ctx := context.Background()

	uuid := uuid.New()
	r.ID = uuid.String()

	bytArr, err := json.Marshal(r)
	if err != nil {
		logger.Print(err.Error())
	}

	bi := &dapr.BindingInvocation{
		Name:      s.components.OutputBindingName,
		Data:      bytArr,
		Operation: "create",
	}

	logger.Printf("invoking binding '%s'", s.components.OutputBindingName)
	err = s.client.InvokeOutputBinding(ctx, bi)
	if err != nil {
		logger.Print(err.Error())
	}

	logger.Printf("new result with RaceID: '%d' & RaceName: '%s'", r.RaceID, r.RaceName)
	return nil
}

// ResultsTopicHandler listens for new results from a ServiceBus Topic
func (s *ResultsService) ResultsTopicHandler(ctx context.Context, e *common.TopicEvent) error {
	defer helper.TimeTrack(time.Now(), "resultsEventHandler()")

	d, err := json.Marshal(e.Data)
	if err != nil {
		logger.Fatal(err)
	}

	var result spec.Result
	err = json.Unmarshal(d, &result)
	if err != nil {
		logger.Fatal(err)
	}

	// save result to result DB
	s.AddResult(result, "results-binding")

	logger.Printf("event - Topic: '%s', RaceID: '%d', RaceName: '%s', TrackName: '%s', Time: '%s'", e.Topic, result.RaceID, result.RaceName, result.TrackName, result.Time)

	for _, r := range result.Runners {
		err := s.ProcessResult(r.HorseID, result.RaceID)
		if err != nil {
			log.Fatalf(err.Error())
		}
	}

	return nil
}

// ProcessResult processes a race result
func (s *ResultsService) ProcessResult(horseID int, raceID int) error {
	defer helper.TimeTrack(time.Now(), "processResult()")
	ctx := context.Background()

	betData := spec.BetData{
		RaceID:  raceID,
		HorseID: horseID,
	}

	bd, err := json.Marshal(betData)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	content := &dapr.DataContent{
		ContentType: "application/json",
		Data:        bd,
	}

	logger.Printf("betResults method call input data: %s", string(bd))
	resp, err := s.client.InvokeServiceWithContent(ctx, "bets", "betresults", content)
	if err != nil {
		logger.Printf("Error invoking method 'betresults' on service 'bets': %s", err.Error())
	}
	if resp != nil {
		logger.Printf("method 'betresults' invoked on service 'bets' with response: %s\n", string(resp))
	} else {
		logger.Print("method 'betresults' invoked on service 'bets' and returned no response")
	}

	return nil
}

// SendSignalRMessage sends a message to Azure Signalr service
func (s *ResultsService) SendSignalRMessage(signalRName string, signalRHubName string) error {
	ctx := context.Background()

	message := spec.SignalRMessage{}
	message.Target = s.components.SignalRHubName
	message.Arguments = []string{"New Message!"}

	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(message)

	out := &dapr.BindingInvocation{
		Name:      s.components.SignalRName,
		Operation: "create",
		Data:      reqBodyBytes.Bytes(),
	}

	logger.Printf("invoking SignalR output binding")
	err := s.client.InvokeOutputBinding(ctx, out)
	if err != nil {
		logger.Printf(err.Error())
		return err
	}

	return nil
}
