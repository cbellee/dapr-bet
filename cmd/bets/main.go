package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cbellee/dapr-bet/cmd/bets/impl"
	"github.com/cbellee/dapr-bet/cmd/bets/spec"
)

var (
	version               = "0.0.1"            // App version number, set at build time with -ldflags "-X 'main.version=1.2.3'"
	buildInfo             = "No build details" // Build details, set at build time with -ldflags "-X 'main.buildInfo=Foo bar'"
	serviceName           = "bets"
	servicePort           = "8004"
	cosmosDbName          = "bets"
	cosmosDbContainerName = "default"
	messageBusName        = "messagebus"
	stateStoreName        = "bets-statestore"
	bindingName           = "bets-binding"
	secretStoreName       = "azurekeyvault"
	topicName             = "payments"
	logger                = log.New(os.Stdout, "", 0)
	cosmosDbKey           = os.Getenv("COSMOS_DB_KEY")
	cosmosDbURL           = os.Getenv("COSMOS_DB_URL")
)

var dbConfig = spec.DbConfig{
	DbURL:       cosmosDbURL,
	DbName:      cosmosDbName,
	DbKey:       cosmosDbKey,
	DbContainer: cosmosDbContainerName,
}

var components = spec.DaprComponents{
	MessageBusName: messageBusName,
	TopicName:      topicName,
	BindingName:    bindingName,
}

// API type
type API struct {
	service spec.BetsService
}

func main() {
	logger.Printf("### Dapr: %v v%v starting...", serviceName, version)

	api := API{
		impl.NewService(serviceName, servicePort, dbConfig, components),
	}

	if err := api.service.AddServiceHandler("/add", api.service.AddBet); err != nil {
		logger.Fatalf("error adding 'AddBet' invocation handler: %v", err)
	}

	if err := api.service.AddServiceHandler("/get", api.service.GetBet); err != nil {
		logger.Fatalf("error adding 'GetBet' invocation handler: %v", err)
	}

	if err := api.service.AddServiceHandler("/getall", api.service.GetBets); err != nil {
		logger.Fatalf("error adding 'GetBets' invocation handler: %v", err)
	}

	if err := api.service.AddServiceHandler("/getbyemail", api.service.GetBetsByEmail); err != nil {
		logger.Fatalf("error adding 'GetBets' invocation handler: %v", err)
	}

	if err := api.service.AddServiceHandler("/betresults", api.service.GetBetResults); err != nil {
		logger.Fatalf("error adding 'GetBets' invocation handler: %v", err)
	}

	if err := api.service.StartService(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("error: %v", err)
	}
}
