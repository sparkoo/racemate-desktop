package acc

import (
	"compress/gzip"
	"encoding/gob"
	"fmt"
	"os"
)

func saveToFile(filename string, data []*Frame) error {
	// Register the struct with gob (important for pointers)
	gob.Register(&Frame{})

	// Create a file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	// Create a gob encoder
	encoder := gob.NewEncoder(gzipWriter)

	// Encode the slice of pointers to structs
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data: %v", err)
	}

	// Ensure data is fully written before closing
	gzipWriter.Flush()

	return nil
}

func loadFromFile(filename string) ([]*Frame, error) {
	// Register the struct with gob
	gob.Register(&Frame{})

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	// Create a gob decoder
	decoder := gob.NewDecoder(gzipReader)

	// Decode into a slice of pointers to structs
	var data []*Frame
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	return data, nil
}
