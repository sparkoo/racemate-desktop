package acc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sparkoo/acctelemetry-go"
	message "github.com/sparkoo/racemate-msg/dist"
)

var currentLap *message.Lap
var lastFrame *message.Frame

var scraping = false

func scrape(ctx context.Context, telemetry *acctelemetry.AccTelemetry) {
	if !scraping {
		fmt.Println("starting scraping the telemetry")
		scraping = true
		ticker := time.NewTicker(10 * time.Millisecond)
		go func(telemetry *acctelemetry.AccTelemetry) {
			currentLap = startNewLap(telemetry)
			for _ = range ticker.C {
				if !scraping {
					ticker.Stop()
				}
				frame := copyToFrame(telemetry)
				processFrame(ctx, frame, telemetry)
				lastFrame = frame
			}
		}(telemetry)
	}
}

func processFrame(ctx context.Context, frame *message.Frame, telemetry *acctelemetry.AccTelemetry) {
	// check if we're in new lap
	if lastFrame != nil && frame.NormalizedCarPosition-lastFrame.NormalizedCarPosition < 0 {
		fmt.Printf("new lap. Is it valid? '%d'\n", lastFrame.IsValidLap)
		if lastFrame.IsValidLap == 1 { // we care only if it is valid lap
			justFinishedLap := currentLap
			justFinishedLap.LapTimeMs = telemetry.GraphicsPointer().ILastTime
			justFinishedLap.Timestamp = uint64(time.Now().Unix())
			go saveToFile(ctx, fmt.Sprintf("%s.%s", strconv.FormatInt(time.Now().Unix(), 10), "lap"), justFinishedLap)
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
	str := ""
	for _, val := range arr {
		if val > 0 {
			str += string(rune(val))
		}
	}
	return strings.TrimSpace(str)
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
