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

	"github.com/sparkoo/racemate-desktop/pkg/auth"
	"github.com/sparkoo/racemate-desktop/pkg/state"
)

func UploadJob(ctx context.Context) error {
	appState, err := state.GetAppState(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get at upload job: %w", err)
	}

	// Create auth manager once
	authManager := auth.NewAuthManager(appState)

	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		// Skip upload if telemetry is online (we're racing)
		if appState.TelemetryOnline {
			continue
		}

		// Check if there are laps to upload
		hasLapsToUpload := hasLapsToUpload(appState)
		if !hasLapsToUpload {
			continue
		}

		// Check if user is authenticated before attempting upload
		if !authManager.IsLoggedIn() {
			// Log this only occasionally to avoid spamming the log
			if time.Now().Second()%30 == 0 {
				fmt.Println("Skipping upload: user not logged in but has laps to upload")
			}
			continue
		}

		// Proceed with upload since user is authenticated and has laps
		if uploadErr := UploadSingleLap(appState); uploadErr != nil {
			fmt.Printf("Failed to upload a single lap: %s\n", uploadErr)
		}
	}

	return nil
}

// hasLapsToUpload checks if there are any lap files waiting to be uploaded
func hasLapsToUpload(appState *state.AppState) bool {
	entries, err := os.ReadDir(appState.UploadDir)
	if err != nil {
		fmt.Printf("Failed to read upload directory: %s\n", err)
		return false
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lap.gzip") {
			return true
		}
	}
	return false
}

func UploadSingleLap(appState *state.AppState) error {
	entries, err := os.ReadDir(appState.UploadDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lap.gzip") {
			fmt.Printf("Uploading '%s'\n", entry.Name())
			lapFile := fmt.Sprintf("%s/%s", appState.UploadDir, entry.Name())
			uploadErr := UploadFile(lapFile, appState)
			if uploadErr != nil {
				return fmt.Errorf("Failed to upload the file: %w", uploadErr)
			}
			fmt.Println("File uploaded successfully")
			err := os.Rename(lapFile, fmt.Sprintf("%s/%s", appState.UploadedDir, entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to move the file '%s' to uploaded directory: %w", entry.Name(), err)
			}
			fmt.Println("File moved to uploaded directory")
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

	// Add authorization token if user is logged in
	authManager := auth.NewAuthManager(appState)
	if authManager.IsLoggedIn() {
		user, err := authManager.GetCurrentUser()
		if err == nil && user != nil {
			req.Header.Set("Authorization", "Bearer "+user.IDToken)
			fmt.Println("Added auth token to upload request")
		} else {
			fmt.Println("Failed to get user for auth token:", err)
		}
	} else {
		fmt.Println("User not logged in, upload will be unauthorized")
	}

	// Send request using http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending upload request: %w", err)
	}
	defer resp.Body.Close()

	// Print response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Upload failed with status code: %d %s", resp.StatusCode, resp.Status)
	}
	fmt.Println("Upload Response Status:", resp.Status)
	return nil
}
