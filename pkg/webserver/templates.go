package webserver

import (
	"embed"
	"encoding/json"
	"log/slog"
)

//go:embed templates/login.html
//go:embed embedded/firebase_config.json
var templateFS embed.FS

// FirebaseConfigJSON holds the embedded Firebase configuration
type FirebaseConfigJSON struct {
	APIKey            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
	AppID             string `json:"appId"`
	MeasurementID     string `json:"measurementId"`
}

// LoadEmbeddedFirebaseConfig loads the Firebase configuration from the embedded file
func LoadEmbeddedFirebaseConfig() (*FirebaseConfigJSON, error) {
	configData, err := templateFS.ReadFile("embedded/firebase_config.json")
	if err != nil {
		slog.Error("Error reading embedded Firebase config", "error", err)
		return nil, err
	}

	var config FirebaseConfigJSON
	err = json.Unmarshal(configData, &config)
	if err != nil {
		slog.Error("Error unmarshaling Firebase config", "error", err)
		return nil, err
	}

	return &config, nil
}
