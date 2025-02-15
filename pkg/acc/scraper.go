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

type Scraper struct {
	currentLap *message.Lap
	lastFrame  *message.Frame

	scraping bool
}

func (s *Scraper) scrape(ctx context.Context, telemetry *acctelemetry.AccTelemetry) {
	if !s.scraping {
		fmt.Println("starting scraping the telemetry")
		s.scraping = true
		go func(telemetry *acctelemetry.AccTelemetry) {
			ticker := time.NewTicker(10 * time.Millisecond)
			s.currentLap = startNewLap(telemetry)
			for _ = range ticker.C {
				if !s.scraping {
					ticker.Stop()
				}
				frame := copyToFrame(telemetry)
				s.processFrame(ctx, frame, telemetry)
				s.lastFrame = frame
			}
		}(telemetry)
	}
}

func (s *Scraper) processFrame(ctx context.Context, frame *message.Frame, telemetry *acctelemetry.AccTelemetry) {
	// check if we're in new lap
	if s.lastFrame != nil && frame.NormalizedCarPosition-s.lastFrame.NormalizedCarPosition < 0 {
		// we care only if it is valid lap
		if s.lastFrame.IsValidLap == 1 && s.currentLap.Frames[0].NormalizedCarPosition < 0.01 {
			justFinishedLap := s.currentLap
			justFinishedLap.LapTimeMs = telemetry.GraphicsPointer().ILastTime
			justFinishedLap.Timestamp = uint64(time.Now().Unix())
			go saveToFile(ctx, fmt.Sprintf("%s_%s_%s.%s", strconv.FormatInt(time.Now().Unix(), 10), justFinishedLap.Track, justFinishedLap.CarModel, "lap"), justFinishedLap)
		}

		s.currentLap = startNewLap(telemetry)
	}
	s.currentLap.Frames = append(s.currentLap.Frames, frame)
}

func (s *Scraper) stop() {
	fmt.Println("stop scraping")
	s.scraping = false
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
