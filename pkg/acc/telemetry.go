package acc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sparkoo/acctelemetry-go"
	"github.com/sparkoo/racemate-desktop/pkg/state"
)

type TelemetryState struct {
	telemetry *acctelemetry.AccTelemetry
}

func TelemetryLoop(ctx context.Context) {
	log := state.GetLogger(ctx)
	telemetry := &TelemetryState{telemetry: acctelemetry.New(acctelemetry.DefaultUdpConfig())}
	scraper := &Scraper{}
	appState, err := state.GetAppState(ctx)
	if err != nil {
		fmt.Printf("failed to get app state in TelemetryLoop: %s", err)
	}

	// this loop is checking whether we have running ACC session
	for range time.NewTicker(10 * time.Second).C {
		if appState.TelemetryOnline {
			if telemetry.telemetry.GraphicsPointer() != nil && telemetry.telemetry.GraphicsPointer().ACStatus != 2 {
				appState.TelemetryOnline = false
			}
		} else {
			if connectionErr := telemetry.telemetry.Connect(); connectionErr == nil {
				if telemetry.telemetry.GraphicsPointer().ACStatus == 2 {
					appState.TelemetryOnline = true
					scraper.scrape(ctx, telemetry.telemetry)
				} else {
					telemetry.telemetry.Close()
					scraper.stop()
				}
			} else {
				log.Error("failed to connect, trying again...", slog.Any("err", connectionErr))
			}
		}
	}
}
