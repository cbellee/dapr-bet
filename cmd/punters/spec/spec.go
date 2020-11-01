package spec

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

// Punter
type Punter struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"firstname"`
	LastName  string  `json:"lastname"`
	Balance   float32 `json:"balance"`
}

type Bet struct {
	ID      string  `json:"id"`
	Email   string  `json:"email"`
	Odds    string  `json:"odds"`
	RaceID  int     `json:"raceid"`
	HorseID int     `json:"horseid"`
	Place   int     `json:"place"`
	Amount  float32 `json:"amount"`
	Paid    bool    `json:"paid"`
}

type CreditDebit struct {
	Amount float32 `json:"amount"`
}

// DbConfig holds cosmosDB configuration/connection info
type DbConfig struct {
	DbURL       string
	DbName      string
	DbKey       string
	DbContainer string
}

// DaprComponents
type DaprComponents struct {
	MessageBusName string `json:"messagebus"`
	StateStoreName string `json:"statestorename"`
}

type PuntersService interface {
	StartService() error
	AddTopicHandler(sub *common.Subscription, fn func(context.Context, *common.TopicEvent) error) error
	AddServiceHandler(name string, fn func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error
	GetPunter(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	AddPunter(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	CalculateBalance(ctx context.Context, in *common.TopicEvent) error
}
