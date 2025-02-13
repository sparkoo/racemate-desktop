package upload

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
)

func UploadFile(filename string) error {
	fileBytes, readFileErr := os.ReadFile(filename)
	if readFileErr != nil {
		return fmt.Errorf("failed to read the file for the upload: %w", readFileErr)
	}

	// Create HTTP request
	url := "https://hello-hwppiybqxq-ey.a.run.app"
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
