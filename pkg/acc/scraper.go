package acc

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sparkoo/acctelemetry-go"
	message "github.com/sparkoo/racemate-msg/proto"
)

var currentLap *message.Lap
var lastFrame *message.Frame

var scraping = false

func scrape(telemetry *acctelemetry.AccTelemetry) {
	if !scraping {
		fmt.Println("starting scraping the telemetry")
		scraping = true
		ticker := time.NewTicker(100 * time.Millisecond)
		go func(telemetry *acctelemetry.AccTelemetry) {
			currentLap = startNewLap(telemetry)
			for _ = range ticker.C {
				if !scraping {
					ticker.Stop()
				}
				frame := copyToFrame(telemetry)
				processFrame(frame, telemetry)
				lastFrame = frame
			}
		}(telemetry)
	}
}

func processFrame(frame *message.Frame, telemetry *acctelemetry.AccTelemetry) {
	// check if we're in new lap
	if lastFrame != nil && frame.NormalizedCarPosition-lastFrame.NormalizedCarPosition < 0 {
		fmt.Printf("new lap. Is it valid? '%d'\n", lastFrame.IsValidLap)
		if lastFrame.IsValidLap == 1 { // we care only if it is valid lap
			justFinishedLap := currentLap
			justFinishedLap.LapTimeMs = telemetry.GraphicsPointer().ILastTime
			go saveToFile(fmt.Sprintf("%s.%s", strconv.FormatInt(time.Now().Unix(), 10), "json"), justFinishedLap)
		}

		currentLap = startNewLap(telemetry)
	}
	currentLap.Frames = append(currentLap.Frames, frame)
}

func startNewLap(telemetry *acctelemetry.AccTelemetry) *message.Lap {
	static := telemetry.StaticPointer()
	physics := telemetry.PhysicsPointer()
	graphics := telemetry.GraphicsPointer()
	return &message.Lap{
		SmVersion:        uint16SliceToString(static.SmVersion[:]),
		AcVersion:        uint16SliceToString(static.AcVersion[:]),
		NumberOfSessions: static.NumberOfSessions,
		CarModel:         uint16SliceToString(static.CarModel[:]),
		Track:            uint16SliceToString(static.Track[:]),
		PlayerName:       uint16SliceToString(static.PlayerName[:]),
		PlayerNick:       uint16SliceToString(static.PlayerNick[:]),
		PlayerSurname:    uint16SliceToString(static.PlayerSurname[:]),
		AirTemp:          physics.AirTemp,
		RoadTemp:         physics.RoadTemp,
		SessionType:      graphics.ACSessionType,
		RainTyres:        graphics.RainTyres,
		IsValidLap:       graphics.IsValidLap,
		TrackGripStatus:  graphics.TrackGripStatus,
		RainIntensity:    graphics.RainIntensity,
		SessionIndex:     graphics.SessionIndex,
		LapTimeMs:        0,
		Frames:           make([]*message.Frame, 1000),
	}
}

func uint16SliceToString(arr []uint16) string {
	var str strings.Builder // Use strings.Builder for efficiency
	for _, val := range arr {
		str.WriteString(strconv.FormatUint(uint64(val), 10)) // Convert to uint64 for FormatUint
		str.WriteString(" ")                                 // Add a space between numbers (optional)
	}
	return strings.TrimSpace(str.String()) // Remove trailing space
}

func stop() {
	scraping = false
}

func copyToFrame(telemetry *acctelemetry.AccTelemetry) *message.Frame {
	// static := telemetry.StaticPointer()
	physics := telemetry.PhysicsPointer()
	graphics := telemetry.GraphicsPointer()

	return &message.Frame{
		GraphicPacket: graphics.PacketID,
		PhysicsPacket: physics.PacketID,
		IsValidLap:    graphics.IsValidLap,

		Gas:        physics.Gas,
		Brake:      physics.Brake,
		Gear:       physics.Gear,
		Rpm:        physics.RPMs,
		SteerAngle: physics.SteerAngle,
		SpeedKmh:   physics.SpeedKmh,

		CurrentTime:           graphics.ICurrentTime,
		NormalizedCarPosition: graphics.NormalizedCarPosition,
		CarCoordinateX:        graphics.CarCoordinates[0][0],
		CarCoordinateY:        graphics.CarCoordinates[0][1],
		CarCoordinateZ:        graphics.CarCoordinates[0][2],
	}
}
