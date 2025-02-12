package acc

import (
	"encoding/json"
	"fmt"
	"os"

	message "github.com/sparkoo/racemate-msg/proto"
	"google.golang.org/protobuf/proto"
)

func saveToFile(filename string, data *message.Lap) error {
	protobufMessage, protoErr := proto.Marshal(data)
	if protoErr != nil {
		return fmt.Errorf("failed to marshal lap message with protobuf: %w", protoErr)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	if _, errWrite := file.Write(protobufMessage); errWrite != nil {
		return fmt.Errorf("failed write to file: %w", errWrite)
	}

	return nil
}

func saveToJson(filename string, lap *message.Lap) error {
	fmt.Printf("Save lap to '%s'\n", filename)
	jsonData, err := json.MarshalIndent(lap, "", "  ") // Use MarshalIndent for pretty printing
	if err != nil {
		return fmt.Errorf("marshaling frames to JSON: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}
