package spec

import (
	"context"

	"github.com/dapr/go-sdk/service/common"
)

// Race struct defines a race
type Race struct {
	RaceID    int    `json:"raceid"`
	RaceName  string `json:"racename"`
	TrackName string `json:"trackname"`
	Time      string `json:"time"`
	Runners   []struct {
		HorseID   int    `json:"horseid"`
		HorseName string `json:"horsename"`
		Odds      string `json:"odds"`
	} `json:"runners"`
}

// DbConfig holds cosmosDB configuration/connection info
type DbConfig struct {
	DbURL       string
	DbName      string
	DbKey       string
	DbContainer string
}

// DaprComponents holds the various Dapr components
type DaprComponents struct {
	StateStoreName string `json:"statestorename"`
}

// RacesService defines the interface
type RacesService interface {
	GetRaces(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)
	AddServiceHandler(name string, fn func(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error)) error
	StartService() error
}
