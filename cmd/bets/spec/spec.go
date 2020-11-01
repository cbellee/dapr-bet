package spec

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

// Bet
type Bet struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	Odds      string  `json:"odds"`
	RaceID    int     `json:"raceid"`
	HorseID   int     `json:"horseid"`
	RaceName  string  `json:"racename"`
	HorseName string  `json:"horsename"`
	Type      string  `json:"type"`
	Amount    float32 `json:"amount"`
	Paid      bool    `json:"paid"`
}

// DbConfig is a struct
type DbConfig struct {
	DbURL       string
	DbName      string
	DbKey       string
	DbContainer string
}

// DaprComponents is a struct
type DaprComponents struct {
	TopicName      string `json:"topicname"`
	MessageBusName string `json:"messagebus"`
	BindingName    string `json:"bindingname"`
}

// BetsService defines the behaviours needed to interact with the service
type BetsService interface {
	AddTopicHandler(sub *common.Subscription, fn func(context.Context, *common.TopicEvent) error) error
	AddServiceHandler(name string, fn func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error
	AddBet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	GetBet(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	GetBets(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	GetBetsByEmail(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	GetBetResults(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	StartService() error
}
