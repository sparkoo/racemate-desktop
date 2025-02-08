package acc

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sparkoo/acctelemetry-go"
)

type Lap struct {
	SmVersion        [15]uint16
	AcVersion        [15]uint16
	NumberOfSessions int32
	CarModel         [33]uint16
	Track            [33]uint16
	PlayerName       [33]uint16
	PlayerNick       [33]uint16
	PlayerSurname    [33]uint16
	AirTemp          float32
	RoadTemp         float32
	SessionType      int32
	RainTyres        int32
	IsValidLap       int32
	TrackGripStatus  int32
	RainIntensity    int32
	SessionIndex     int32
}

type Frame struct {
	GraphicPacket int32
	PhysicsPacket int32

	Gas                   float32
	Brake                 float32
	Gear                  int32
	RPM                   int32
	SteerAngle            float32
	SpeedKmh              float32
	ICurrentTime          int32
	NormalizedCarPosition float32
	CarCoordinates        [3]float32
	// TCLevel               int32
	// TCCutLevel            int32
	// EngineMapLevel        int32
	// ABSLevel              int32
}

var scraping = false
var scraped []*Frame

func scrape(telemetry *acctelemetry.AccTelemetry) {
	if !scraping {
		fmt.Println("starting scraping the telemetry")
		scraping = true
		ticker := time.NewTicker(10 * time.Millisecond)
		go func(telemetry *acctelemetry.AccTelemetry) {
			for _ = range ticker.C {
				if !scraping {
					ticker.Stop()
				}
				scraped = append(scraped, copyToFrame(telemetry))
			}
		}(telemetry)
	}
}

func stop() {
	scraping = false
	timestamp := fmt.Sprintf("%s.%s", strconv.FormatInt(time.Now().Unix(), 10), "gob")
	fmt.Printf("stopping scraping the telemetry. Lets flush it to the file '%s'\n", timestamp)
	// fmt.Printf("So what do we have here:")
	// for _, f := range scraped {
	// 	fmt.Printf("frame: %+v\n", f)
	// }
	saveToFile(timestamp, scraped)
	scraped = make([]*Frame, 0)
}

func copyToFrame(telemetry *acctelemetry.AccTelemetry) *Frame {
	// static := telemetry.StaticPointer()
	physics := telemetry.PhysicsPointer()
	graphics := telemetry.GraphicsPointer()

	return &Frame{
		GraphicPacket: graphics.PacketID,
		PhysicsPacket: physics.PacketID,

		// SmVersion:        static.SmVersion,
		// AcVersion:        static.AcVersion,
		// NumberOfSessions: static.NumberOfSessions,
		// CarModel:         static.CarModel,
		// Track:            static.Track,
		// PlayerName:       static.PlayerName,
		// PlayerNick:       static.PlayerNick,
		// PlayerSurname:    static.PlayerSurname,

		Gas:        physics.Gas,
		Brake:      physics.Brake,
		Gear:       physics.Gear,
		RPM:        physics.RPMs,
		SteerAngle: physics.SteerAngle,
		SpeedKmh:   physics.SpeedKmh,
		// WheelPressure:      physics.WheelsPressure,
		// TyreCoreTemp:       physics.TyreCoreTemperature,
		// TC:                 physics.TC,
		// ABS:                physics.ABS,
		// CarDamage:          physics.CarDamage,
		// PitLimiter:         physics.PitLimiterOn,
		// AirTemp:            physics.AirTemp,
		// RoadTemp:           physics.RoadTemp,
		// FinalFF:            physics.FinalFF,
		// BrakeTemp:          physics.BrakeTemp,
		// IsAIController:     physics.IsAIControlled,
		// FrontBrakeCompound: physics.FrontBrakeCompound,
		// RearBrakeCompound:  physics.RearBrakeCompound,
		// PadLife:            physics.PadLife,
		// DiscLife:           physics.DiscLife,

		// SessionType:           graphics.ACSessionType,
		// CompletedLaps:         graphics.CompletedLaps,
		// Position:              graphics.Position,
		ICurrentTime: graphics.ICurrentTime,
		// ILastTime:             graphics.ILastTime,
		// IBestTime:             graphics.IBestTime,
		NormalizedCarPosition: graphics.NormalizedCarPosition,
		CarCoordinates:        graphics.CarCoordinates[0],
		// PlayerCarID:           graphics.PlayerCarID,
		// WindSpeed:             graphics.WindSpeed,
		// WindDirection:         graphics.WindDirection,
		// TCLevel:               graphics.TC,
		// TCCutLevel:            graphics.TCCut,
		// EngineMapLevel:        graphics.EngineMap,
		// ABSLevel:              graphics.ABS,
		// FuelXLap:              graphics.FuelXLap,
		// RainTyres:             graphics.RainTyres,
		// SessionIndex:          graphics.SessionIndex,
		// IsValidLap:            graphics.IsValidLap,
		// TrackStatus:           graphics.TrackStatus,
		// Clock:                 graphics.Clock,
		// TrackGripStatus:       graphics.TrackGripStatus,
		// RainIntensity:         graphics.RainIntensity,
	}
}
