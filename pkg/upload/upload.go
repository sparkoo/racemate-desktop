package upload

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sparkoo/racemate-desktop/pkg/state"
)

func UploadJob(ctx context.Context) error {
	appState, err := state.GetAppState(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get at upload job: %w", err)
	}

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		if !appState.TelemetryOnline {
			UploadSingleLap(appState)
		}
	}

	return nil
}

func UploadSingleLap(appState *state.AppState) error {
	entries, err := os.ReadDir(appState.UploadDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lap.gzip") {
			fmt.Printf("Uploading '%s'", entry.Name())
			lapFile := fmt.Sprintf("%s/%s", appState.UploadDir, entry.Name())
			uploadErr := UploadFile(lapFile, appState)
			if uploadErr != nil {
				return fmt.Errorf("Failed to upload the file: %w", uploadErr)
			}
			err := os.Rename(lapFile, fmt.Sprintf("%s/%s", appState.UploadedDir, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to move the file '%s' to uploaded directory: %w", entry.Name(), err)
			}
			break
		}
	}
	return nil
}

func UploadFile(filename string, appState *state.AppState) error {
	fileBytes, readFileErr := os.ReadFile(filename)
	if readFileErr != nil {
		return fmt.Errorf("failed to read the file for the upload: %w", readFileErr)
	}

	// Create HTTP request
	url := appState.UploadURL
	req, err := http.NewRequest("POST", url, bytes.NewReader(fileBytes))
	if err != nil {
		return fmt.Errorf("Error creating upload request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("content-encoding", "gzip")

	// Send request using http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending upload request: %w", err)
	}
	defer resp.Body.Close()

	// Print response
	fmt.Println("Upload Response Status:", resp.Status)
	return nil
}
