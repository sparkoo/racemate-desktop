package acc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sparkoo/acctelemetry-go"
	"github.com/sparkoo/racemate-desktop/pkg/state"
	message "github.com/sparkoo/racemate-msg/dist"
)

type Scraper struct {
	currentLap *message.Lap
	lastFrame  *message.Frame

	scraping bool
}

func (s *Scraper) scrape(ctx context.Context, telemetry *acctelemetry.AccTelemetry) {
	appState, err := state.GetAppState(ctx)
	var pollRate time.Duration
	if err != nil {
		fmt.Printf("failed to get app state to save to the file: %s", err)
		pollRate = 1 * time.Second
	} else {
		pollRate = appState.PollRate
	}

	if !s.scraping {
		fmt.Println("starting scraping the telemetry")
		s.scraping = true
		go func(telemetry *acctelemetry.AccTelemetry) {
			ticker := time.NewTicker(pollRate) // main ticker for polling the telemetry data
			s.currentLap = startNewLap(telemetry)
			for _ = range ticker.C {
				if !s.scraping {
					ticker.Stop()
				}
				frame := copyToFrame(telemetry)
				if frame != nil {
					s.processFrame(ctx, frame, telemetry)
					s.lastFrame = frame
				}
			}
		}(telemetry)
	}
}

func (s *Scraper) processFrame(ctx context.Context, frame *message.Frame, telemetry *acctelemetry.AccTelemetry) {
	// check if we're in new lap
	if len(s.currentLap.Frames) > 0 && s.lastFrame != nil && frame.NormalizedCarPosition-s.lastFrame.NormalizedCarPosition < 0 {
		// we care only if it is valid lap
		firstFrame := s.currentLap.Frames[0]
		lastFrame := s.currentLap.Frames[len(s.currentLap.Frames)-1]
		if s.lastFrame.IsValidLap == 1 && firstFrame.NormalizedCarPosition < 0.05 && lastFrame.NormalizedCarPosition > 0.95 {
			justFinishedLap := s.currentLap
			justFinishedLap.LapTimeMs = telemetry.GraphicsPointer().ILastTime
			justFinishedLap.Timestamp = uint64(time.Now().Unix())
			go s.finalizeLap(ctx, justFinishedLap, telemetry)

		} else {
			fmt.Printf("lap is not valid, %d, %f\n", s.lastFrame.IsValidLap, s.currentLap.Frames[0].NormalizedCarPosition)
		}

		s.currentLap = startNewLap(telemetry)
	}
	s.currentLap.Frames = append(s.currentLap.Frames, frame)
}

func (s *Scraper) finalizeLap(ctx context.Context, lap *message.Lap, telemetry *acctelemetry.AccTelemetry) {
	// find car update from UDP, let's try 10s
	start := time.Now()
	tries := 0
	for time.Since(start) < 10*time.Second {
		tries += 1
		carUpdateMessage := telemetry.ReadUdpMessage()
		//TODO: check if we're in current lap already
		if carUpdateMessage != nil && lap.CarIndex == int32(carUpdateMessage.CarIndex) {
			fmt.Printf("We've found it in %d tries \\o/ is it valid? %+v\n", tries, carUpdateMessage)
			fmt.Printf("current: %+v\nbest: %+v\nlast: %+v\n", carUpdateMessage.CurrentLap, carUpdateMessage.BestSessionLap, carUpdateMessage.LastLap)
			if carUpdateMessage.LastLap.InValidForBest > 0 {
				fmt.Println("it is valid")
				saveToFile(ctx, fmt.Sprintf("%s_%s_%s.%s", strconv.FormatInt(time.Now().Unix(), 10), lap.Track, lap.CarModel, "lap"), lap)
			} else {
				fmt.Println("it is not valid")
			}
			return
		}

		time.Sleep(100 * time.Millisecond)
	}

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
		SmVersion:       uint16SliceToString(static.SmVersion[:]),
		AcVersion:       uint16SliceToString(static.AcVersion[:]),
		CarModel:        uint16SliceToString(static.CarModel[:]),
		Track:           uint16SliceToString(static.Track[:]),
		PlayerName:      uint16SliceToString(static.PlayerName[:]),
		PlayerNick:      uint16SliceToString(static.PlayerNick[:]),
		PlayerSurname:   uint16SliceToString(static.PlayerSurname[:]),
		AirTemp:         physics.AirTemp,
		RoadTemp:        physics.RoadTemp,
		CarIndex:        graphics.PlayerCarID,
		SessionType:     graphics.ACSessionType,
		RainTyres:       graphics.RainTyres,
		IsValidLap:      graphics.IsValidLap,
		TrackGripStatus: graphics.TrackGripStatus,
		RainIntensity:   graphics.RainIntensity,
		LapTimeMs:       0,
		Frames:          make([]*message.Frame, 0),
		LapNumber:       graphics.NumberOfLaps,
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
	if graphics == nil || physics == nil {
		return nil
	}

	// find out on what array ID is my car
	// we have to do this for every frame as as players disconnects, my index might change
	// it may be an issue if someone disconnects after we get the index and when we actually read the data, then we might get wrong coordinates
	// but it may be so rare, that it will never happen
	// let's fix once it is real issue
	carIndex := 0
	for _, carId := range graphics.CarID {
		if carId == graphics.PlayerCarID {
			break
		}
		carIndex++
	}

	return &message.Frame{
		GraphicPacket: graphics.PacketID,
		PhysicsPacket: physics.PacketID,
		IsValidLap:    graphics.IsValidLap,
		PenaltyType:   graphics.Penalty,

		Gas:        physics.Gas,
		Brake:      physics.Brake,
		Gear:       physics.Gear,
		Rpm:        physics.RPMs,
		SteerAngle: physics.SteerAngle,
		SpeedKmh:   physics.SpeedKmh,

		CurrentTime:           graphics.ICurrentTime,
		NormalizedCarPosition: graphics.NormalizedCarPosition,
		CarCoordinateX:        graphics.CarCoordinates[carIndex][0],
		CarCoordinateY:        graphics.CarCoordinates[carIndex][1],
		CarCoordinateZ:        graphics.CarCoordinates[carIndex][2],
	}
}
