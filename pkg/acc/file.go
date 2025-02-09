package acc

import (
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
)

func saveToGob(filename string, data []*Frame) error {
	gob.Register(&Frame{})

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	encoder := gob.NewEncoder(gzipWriter)

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %v", err)
	}

	gzipWriter.Flush()

	return nil
}

func loadFromGob(filename string) ([]*Frame, error) {
	gob.Register(&Frame{})

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	decoder := gob.NewDecoder(gzipReader)

	var data []*Frame
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	return data, nil
}

func saveToJson(filename string, lap *Lap) error {
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

func loadFromJson(filename string) ([]*Frame, error) {
	var frames []*Frame

	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	err = json.Unmarshal(jsonData, &frames)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling JSON to frames: %w", err)
	}

	return frames, nil
}
