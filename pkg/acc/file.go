package acc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sparkoo/racemate-desktop/pkg/state"
	message "github.com/sparkoo/racemate-msg/dist"
	"google.golang.org/protobuf/proto"
)

func saveToFile(ctx context.Context, filename string, data *message.Lap) error {
	log := state.GetLogger(ctx)
	log.Info("Saving lap to file")
	protobufMessage, protoErr := proto.Marshal(data)
	if protoErr != nil {
		return fmt.Errorf("failed to marshal lap message with protobuf: %w", protoErr)
	}

	appState, err := state.GetAppState(ctx)
	if err != nil {
		return fmt.Errorf("failed to get app state to save to the file: %w", err)
	}
	filePath := filepath.Join(appState.UploadDir, filename)

	return saveCompressed(filePath, protobufMessage)
}

func loadFromFile(filename string) (*message.Lap, error) {
	data, _ := os.ReadFile(filename)

	lap := &message.Lap{}
	if err := proto.Unmarshal(data, lap); err != nil {
		return nil, fmt.Errorf("failed to Unmarshal the data: %w", err)
	}

	return lap, nil
}

func loadFromFileCompressed(filename string) (*message.Lap, error) {
	// 1. Open the compressed file
	f, err := os.Open(filename) // Replace with your file name
	if err != nil {
		slog.Error("Error opening file", "error", err)
		return nil, fmt.Errorf("failed to open the file to read: %w", err)
	}
	defer f.Close() // Important: Close the file when done

	// 2. Create a gzip reader
	gr, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("Error creating gzip reader: %w", err)
	}
	defer gr.Close() // Important: Close the gzip reader

	// 3. Read and uncompress the data
	var data bytes.Buffer       // Or []byte if you know the size beforehand
	_, err = io.Copy(&data, gr) // Efficiently copies from reader to buffer
	if err != nil {

		return nil, fmt.Errorf("Error uncompressing data: %w", err)
	}

	// If you need the data as a byte slice:
	uncompressedData := data.Bytes()

	lap := &message.Lap{}
	if err := proto.Unmarshal(uncompressedData, lap); err != nil {
		return nil, fmt.Errorf("failed to Unmarshal the data: %w", err)
	}

	return lap, nil
}

func saveCompressed(filename string, data []byte) error {
	compressedFilename := filename + ".gzip"
	f, err := os.Create(compressedFilename)
	if err != nil {
		return fmt.Errorf("failed to create compressed file: %w", err)
	}
	defer f.Close()

	w := gzip.NewWriter(f)
	defer w.Close()

	_, err = w.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write compressed data: %w", err)
	}

	err = w.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush compressed data: %w", err)
	}

	return nil
}

func saveToJson(filename string, lap *message.Lap) error {
	slog.Info("Saving lap to JSON", "filename", filename)
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
