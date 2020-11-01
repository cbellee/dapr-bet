package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cbellee/dapr-bet/cmd/races/spec"
	"github.com/cbellee/dapr-bet/pkg/helper"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
	daprd "github.com/dapr/go-sdk/service/http"
	"github.com/pkg/errors"
	"github.com/vippsas/go-cosmosdb/cosmosapi"
)

var (
	logger = log.New(os.Stdout, "", 0)
)

// RacesService implements a dapr service and client
type RacesService struct {
	client     dapr.Client
	server     common.Service
	dbConfig   spec.DbConfig
	components spec.DaprComponents
}

// NewService creates a new instance of the RacesService
func NewService(serviceName string, servicePort string, dbConfig spec.DbConfig, components spec.DaprComponents) *RacesService {
	client, err := dapr.NewClient()
	if err != nil {
		logger.Panicf("Failed to create Dapr client: %s", err)
		return nil
	}

	port := fmt.Sprintf(":%s", servicePort)
	server := daprd.NewService(port)

	service := &RacesService{
		client,
		server,
		dbConfig,
		components,
	}

	return service
}

// StartService starts the http server
func (s *RacesService) StartService() error {
	err := s.server.Start()
	if err != nil {
		return err
	}
	return nil
}

// AddTopicHandler wires up a new topic event handler
func (s *RacesService) AddTopicHandler(sub *common.Subscription, handler func(context.Context, *common.TopicEvent) error) error {
	err := s.server.AddTopicEventHandler(sub, handler)
	if err != nil {
		logger.Fatalf("Error adding topic event handler: %s", err)
		return err
	}

	return nil
}

// AddServiceHandler wires up a new service invocation handler
func (s *RacesService) AddServiceHandler(name string, handler func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error {
	err := s.server.AddServiceInvocationHandler(name, handler)
	if err != nil {
		logger.Printf("Error registering service handler: %s", err)
		return err
	}
	return nil
}

// GetRaces lists all races in the database
func (s *RacesService) GetRaces(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "GetRacesHandler()")

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

	var races []spec.Race
	_, err = client.ListDocuments(context.Background(), s.dbConfig.DbName, s.dbConfig.DbContainer, &listOps, &races)
	if err != nil {
		err = errors.WithStack(err)
		logger.Fatalf("error in listResponse: %s", err.Error())
		return nil, err
	}

	bytArr, err := json.Marshal(races)
	if err != nil {
		logger.Print(err.Error())
	}

	out = &common.Content{
		Data:        bytArr,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}

	return out, nil
}
