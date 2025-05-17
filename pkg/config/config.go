package config

// FirebaseConfig holds all Firebase configuration values
type FirebaseConfig struct {
    APIKey           string
    AuthDomain       string
    ProjectID        string
    StorageBucket    string
    MessagingSenderID string
    AppID           string
    MeasurementID    string
}

// TemplateData returns the config as a map for template execution
func (c *FirebaseConfig) TemplateData() map[string]string {
    return map[string]string{
        "FirebaseAPIKey":           c.APIKey,
        "FirebaseAuthDomain":       c.AuthDomain,
        "FirebaseProjectID":        c.ProjectID,
        "FirebaseStorageBucket":    c.StorageBucket,
        "FirebaseMessagingSenderID": c.MessagingSenderID,
        "FirebaseAppID":           c.AppID,
        "FirebaseMeasurementID":    c.MeasurementID,
    }
}

// NewFirebaseConfig creates a new FirebaseConfig with the provided values
func NewFirebaseConfig(apiKey, authDomain, projectID, storageBucket, messagingSenderID, appID, measurementID string) *FirebaseConfig {
    return &FirebaseConfig{
        APIKey:           apiKey,
        AuthDomain:       authDomain,
        ProjectID:        projectID,
        StorageBucket:    storageBucket,
        MessagingSenderID: messagingSenderID,
        AppID:           appID,
        MeasurementID:    measurementID,
    }
}
