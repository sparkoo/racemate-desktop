package acc

import (
	"time"

	"github.com/sparkoo/acctelemetry-go"
)

type TelemetryState struct {
	telemetry *acctelemetry.AccTelemetry
	onUpdate  func(*TelemetryState)

	Online bool
}

func TelemetryLoop(onUpdate func(*TelemetryState)) {
	state := &TelemetryState{telemetry: acctelemetry.New(), onUpdate: onUpdate, Online: false}
	onUpdate(state)

	// this loop is checking whether we have running ACC session
	for range time.NewTicker(5 * time.Second).C {
		if state.Online {
			if state.telemetry.GraphicsPointer().ACStatus != 2 {
				state.changeOnline(false)
			}
		} else {
			if state.telemetry.Connect() == nil {
				if state.telemetry.GraphicsPointer().ACStatus == 2 {
					state.changeOnline(true)
				} else {
					state.telemetry.Close()
				}
			}
		}
	}
}

func (s *TelemetryState) changeOnline(online bool) {
	if online != s.Online {
		s.Online = online
		s.onUpdate(s)
	}
}
