package acc

import (
	"context"
	"fmt"
	"time"

	"github.com/sparkoo/acctelemetry-go"
	"github.com/sparkoo/racemate-desktop/pkg/state"
)

type TelemetryState struct {
	telemetry *acctelemetry.AccTelemetry
}

func TelemetryLoop(ctx context.Context) {
	telemetry := &TelemetryState{telemetry: acctelemetry.New()}
	scraper := &Scraper{}
	appState, err := state.GetAppState(ctx)
	if err != nil {
		fmt.Printf("failed to get app state in TelemetryLoop: %s", err)
	}

	// this loop is checking whether we have running ACC session
	for range time.NewTicker(5 * time.Second).C {
		if appState.TelemetryOnline {
			if telemetry.telemetry.GraphicsPointer() != nil && telemetry.telemetry.GraphicsPointer().ACStatus != 2 {
				appState.TelemetryOnline = false
			}
		} else {
			if telemetry.telemetry.Connect() == nil {
				if telemetry.telemetry.GraphicsPointer().ACStatus == 2 {
					appState.TelemetryOnline = true
					scraper.scrape(ctx, telemetry.telemetry)
				} else {
					telemetry.telemetry.Close()
					scraper.stop()
				}
			}
		}
	}
}
