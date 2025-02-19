package state

import (
	"context"
	"fmt"
	"time"
)

type raceMateContextKey string

const APP_STATE = raceMateContextKey("appState")

type AppState struct {
	TelemetryOnline bool
	DataDir         string
	UploadDir       string
	UploadedDir     string
	Error           error
	PollRate        time.Duration
}

func GetAppState(ctx context.Context) (*AppState, error) {
	appState, ok := ctx.Value(APP_STATE).(*AppState)
	if !ok {
		return nil, fmt.Errorf("Failed to get app state from the context")
	}
	return appState, nil
}
