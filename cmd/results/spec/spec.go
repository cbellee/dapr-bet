package spec

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

// Result stores a result
type Result struct {
	ID        string `json:"id"`
	RaceID    int    `json:"raceid"`
	RaceName  string `json:"racename"`
	TrackName string `json:"trackname"`
	Time      string `json:"time"`
	Runners   []struct {
		Place     int    `json:"place"`
		Odds      string `json:"odds"`
		HorseName string `json:"horsename"`
		HorseID   int    `json:"horseid"`
	} `json:"runners"`
}

// PlacedResult stores info for a result
type PlacedResult struct {
	RaceID    int    `json:"raceid"`
	HorseID   int    `json:"horseid"`
	RaceName  string `json:"racename"`
	HorseName string `json:"horsename"`
	Place     int    `json:"place"`
	Odds      string `json:"odds"`
}

// BetData stores info for a bet
type BetData struct {
	RaceID  int
	HorseID int
}

// Bet stores a bet
type Bet struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Odds      string `json:"odds"`
	RaceID    int    `json:"raceid"`
	HorseID   int    `json:"horseid"`
	HorseName string `json:"horsename"`
	Type      string `json:"type"`
	Place     int    `json:"place"`
	Paid      bool   `json:"paid"`
}

// DbConfig holds cosmosDB configuration/connection info
type DbConfig struct {
	DbURL       string
	DbName      string
	DbKey       string
	DbContainer string
}

// SignalRMessage represents a signalr message type
type SignalRMessage struct {
	Target    string   `json:"target"`
	Arguments []string `json:"arguments"`
}

// DaprComponents is a struct
type DaprComponents struct {
	TopicName         string `json:"topicname"`
	MessageBusName    string `json:"messagebus"`
	OutputBindingName string `json:"outputbinding"`
	SignalRName       string `json:"signalr"`
	SignalRHubName    string `json:"signalrhub"`
}

// ResultsService interface
type ResultsService interface {
	GetResults(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	AddResult(r Result, bindingName string) error
	AddTopicHandler(sub *common.Subscription, fn func(context.Context, *common.TopicEvent) error) error
	AddServiceHandler(name string, fn func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error
	ResultsTopicHandler(ctx context.Context, e *common.TopicEvent) error
	ProcessResult(horseID int, RaceID int) error
	StartService() error
}
