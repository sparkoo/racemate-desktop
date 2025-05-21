package state

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type raceMateContextKey string
type loggerKey struct{}

const APP_STATE = raceMateContextKey("appState")

type AppState struct {
	TelemetryOnline bool
	DataDir         string
	UploadDir       string
	UploadedDir     string
	LogsDir         string
	Error           error
	PollRate        time.Duration
	Logger          *slog.Logger
	UploadURL       string
}

func GetAppState(ctx context.Context) (*AppState, error) {
	appState, ok := ctx.Value(APP_STATE).(*AppState)
	if !ok {
		return nil, fmt.Errorf("Failed to get app state from the context")
	}
	return appState, nil
}

func GetLogger(ctx context.Context) *slog.Logger {
	appState, _ := GetAppState(ctx)
	return appState.Logger
}
