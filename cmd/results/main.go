package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cbellee/dapr-bet/cmd/results/impl"
	"github.com/cbellee/dapr-bet/cmd/results/spec"
	"github.com/dapr/go-sdk/service/common"
)

var (
	version               = "0.1.0"            // App version number, set at build time with -ldflags "-X 'main.version=1.2.3'"
	buildInfo             = "No build details" // Build details, set at build time with -ldflags "-X 'main.buildInfo=Foo bar'"
	serviceName           = "results"
	servicePort           = "8003"
	cosmosDbName          = "results"
	cosmosDbContainerName = "default"
	pubSubName            = "messagebus"
	bindingName           = "results-binding"
	topicName             = "results"
	signalRName           = "signalr-output-binding"
	signalRHubName        = "daprbet"
	cosmosDbKey           = os.Getenv("COSMOS_DB_KEY")
	cosmosDbURL           = os.Getenv("COSMOS_DB_URL")
	logger                = log.New(os.Stdout, "", 0)
)

var sub = &common.Subscription{
	PubsubName: pubSubName,
	Topic:      topicName,
	Route:      "/",
}

var dbConfig = spec.DbConfig{
	DbURL:       cosmosDbURL,
	DbName:      cosmosDbName,
	DbKey:       cosmosDbKey,
	DbContainer: cosmosDbContainerName,
}

var components = spec.DaprComponents{
	OutputBindingName: bindingName,
	MessageBusName:    pubSubName,
	TopicName:         topicName,
	SignalRName:       signalRName,
	SignalRHubName:    signalRHubName,
}

// API type
type API struct {
	service spec.ResultsService
}

func main() {
	logger.Printf("### Dapr: %v v%v starting on port %s...", serviceName, version, servicePort)

	api := API{
		impl.NewService(serviceName, servicePort, dbConfig, components),
	}

	if err := api.service.AddTopicHandler(sub, api.service.ResultsTopicHandler); err != nil {
		logger.Fatalf("error adding topic subscription: %v", err)
	}

	if err := api.service.AddServiceHandler("/get", api.service.GetResults); err != nil {
		logger.Fatalf("error adding 'getResultsHandler' invocation handler: %v", err)
	}

	if err := api.service.StartService(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("error: %v", err)
	}
}
