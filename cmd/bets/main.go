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

	/* 	if err := api.service.AddTopicHandler(sub, api.service.ResultsTopicHandler); err != nil {
	   		logger.Fatalf("error adding 'AddBet' invocation handler: %v", err)
	   	}
	*/

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

	/* 	if err := s.AddServiceInvocationHandler("/add", addBetHandler); err != nil {
	   		logger.Fatalf("error adding 'addBetHandler' invocation handler: %v", err)
	   	}

	   	if err := s.AddServiceInvocationHandler("/get", getBetHandler); err != nil {
	   		logger.Fatalf("error adding 'getBetHandler' invocation handler: %v", err)
	   	}

	   	if err := s.AddServiceInvocationHandler("/getall", getBetsHandler); err != nil {
	   		logger.Fatalf("error adding 'getBetsHandler' invocation handler: %v", err)
	   	}

	   	if err := s.AddServiceInvocationHandler("/getbyemail", getBetsByEmailHandler); err != nil {
	   		logger.Fatalf("error adding 'getBetsHandler' invocation handler: %v", err)
	   	}

	   	if err := s.AddServiceInvocationHandler("/betresults", betResultsHandler); err != nil {
	   		logger.Fatalf("error adding 'betResultsHandler' invocation handler: %v", err)
	   	} */
}

/* func getSecret(secretStoreName string, secretName string) (s string, err error) {
	ctx := context.Background()
	client, err := dapr.NewClient()
	if err != nil {
		logger.Printf(err.Error())
		return "", err
	}

	secret, err := client.GetSecret(ctx, secretStoreName, secretName, nil)
	if err != nil {
		logger.Printf("error retrieving secret '%s': %s", secretName, err.Error())
		return "", err
	}

	s = secret[secretName]
	return s, nil
} */

/* // addBetHandler adds a new bet to the database
// input: spec.Bet
func addBetHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "addBetHandler()")

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
		Name:      bindingName,
		Data:      bytArr,
		Operation: "create",
	}

	logger.Printf("invoking binding '%s'", bindingName)
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

// getBetHandler returns a bet from the database
// input: spec.Bet
func getBetHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "getBetHandler()")

	var bet spec.Bet
	err = json.Unmarshal(in.Data, &bet)
	if err != nil {
		logger.Print(err.Error())
	}

	cosmosCfg := cosmosapi.Config{
		MasterKey: dbConfig.DbKey,
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

	client := cosmosapi.New(dbConfig.DbUrl, cosmosCfg, nil, nil)
	queryOps := cosmosapi.DefaultQueryDocumentOptions()
	queryOps.IsQuery = true

	var bets []spec.Bet
	queryResponse, err := client.QueryDocuments(context.Background(), dbConfig.DbName, dbConfig.DbContainer, qry, &bets, queryOps)
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

// getBetsHandler returns all bets for the current result from the database
// input: spec.Bet
func getBetsHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "getBetsHandler()")

	cosmosCfg := cosmosapi.Config{
		MasterKey: dbConfig.DbKey,
	}

	client := cosmosapi.New(dbConfig.DbUrl, cosmosCfg, nil, nil)
	listOps := cosmosapi.ListDocumentsOptions{
		MaxItemCount:        1000,
		AIM:                 "",
		Continuation:        "",
		IfNoneMatch:         "",
		PartitionKeyRangeId: "",
	}

	var bets []spec.Bet
	_, err = client.ListDocuments(context.Background(), dbConfig.DbName, dbConfig.DbContainer, &listOps, &bets)
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

// getBetsByEmailHandler returns all bets for a punter from the database
func getBetsByEmailHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "getBetsByEmailHandler()")

	if in.QueryString["email"] == "" {
		logger.Print("querystring 'email' not provided or nil value, exiting...")
		return
	}

	var email = in.QueryString["email"]
	logger.Printf("QueryString: " + email)

	cosmosCfg := cosmosapi.Config{
		MasterKey: dbConfig.DbKey,
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

	client := cosmosapi.New(dbConfig.DbUrl, cosmosCfg, nil, nil)
	queryOps := cosmosapi.DefaultQueryDocumentOptions()
	queryOps.EnableCrossPartition = true
	queryOps.IsQuery = true

	var bets []spec.Bet
	queryResponse, err := client.QueryDocuments(context.Background(), dbConfig.DbName, dbConfig.DbContainer, qry, &bets, queryOps)
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

func betResultsHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	defer helper.TimeTrack(time.Now(), "betResultsHandler()")

	if in == nil {
		return
	}

	cosmosCfg := cosmosapi.Config{
		MasterKey: dbConfig.DbKey,
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

	client := cosmosapi.New(dbConfig.DbUrl, cosmosCfg, nil, nil)
	queryOps := cosmosapi.DefaultQueryDocumentOptions()
	queryOps.EnableCrossPartition = true
	queryOps.IsQuery = true

	var bets []spec.Bet
	queryResponse, err := client.QueryDocuments(context.Background(), dbConfig.DbName, dbConfig.DbContainer, qry, &bets, queryOps)
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
			sendMessageToTopic(messageBusName, topicName, b)
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
*/
