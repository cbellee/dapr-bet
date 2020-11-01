package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cbellee/dapr-bet/cmd/races/impl"
	"github.com/cbellee/dapr-bet/cmd/races/spec"
)

var (
	version               = "0.1.0"            // App version number, set at build time with -ldflags "-X 'main.version=1.2.3'"
	buildInfo             = "No build details" // Build details, set at build time with -ldflags "-X 'main.buildInfo=Foo bar'"
	serviceName           = "races"
	servicePort           = "8005"
	cosmosDbName          = "races"
	cosmosDbContainerName = "default"
	stateStoreName        = "bets-statestore"
	cosmosDbKey           = os.Getenv("COSMOS_DB_KEY")
	cosmosDbURL           = os.Getenv("COSMOS_DB_URL")
	logger                = log.New(os.Stdout, "", 0)
)

var dbConfig = spec.DbConfig{
	DbURL:       cosmosDbURL,
	DbName:      cosmosDbName,
	DbKey:       cosmosDbKey,
	DbContainer: cosmosDbContainerName,
}

var components = spec.DaprComponents{
	StateStoreName: stateStoreName,
}

// API type
type API struct {
	service spec.RacesService
}

func main() {
	logger.Printf("### Dapr: %v v%v starting...", serviceName, version)

	api := API{
		impl.NewService(serviceName, servicePort, dbConfig, components),
	}

	if err := api.service.AddServiceHandler("/get", api.service.GetRaces); err != nil {
		logger.Fatalf("error adding 'getResultsHandler' invocation handler: %v", err)
	}

	if err := api.service.StartService(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("error: %v", err)
	}
}
