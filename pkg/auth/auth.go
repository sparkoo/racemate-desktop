package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sparkoo/racemate-desktop/pkg/state"
)

// UserData represents the persistent user authentication data
type UserData struct {
	UID           string    `json:"uid"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"displayName"`
	PhotoURL      string    `json:"photoURL"`
	IDToken       string    `json:"idToken"`
	RefreshToken  string    `json:"refreshToken"`
	ExpiresAt     time.Time `json:"expiresAt"`
	LastLoginTime time.Time `json:"lastLoginTime"`
}

// AuthManager handles authentication state persistence
type AuthManager struct {
	appState *state.AppState
	userData *UserData
}

// NewAuthManager creates a new auth manager
func NewAuthManager(appState *state.AppState) *AuthManager {
	return &AuthManager{
		appState: appState,
	}
}

// SaveUserData persists user authentication data to disk
func (am *AuthManager) SaveUserData(userData *UserData) error {
	// Set last login time
	userData.LastLoginTime = time.Now()

	// Store user data in memory
	am.userData = userData

	// Create auth directory if it doesn't exist
	authDir := filepath.Join(am.appState.DataDir, "auth")
	if err := os.MkdirAll(authDir, 0700); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	// Write user data to file
	userDataFile := filepath.Join(authDir, "user.json")
	data, err := json.MarshalIndent(userData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	// Write with restrictive permissions (only user can read/write)
	if err := os.WriteFile(userDataFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write user data file: %w", err)
	}

	return nil
}

// LoadUserData loads user authentication data from disk
func (am *AuthManager) LoadUserData() (*UserData, error) {
	// If already loaded in memory, return it
	if am.userData != nil {
		return am.userData, nil
	}

	// Try to load from file
	userDataFile := filepath.Join(am.appState.DataDir, "auth", "user.json")
	data, err := os.ReadFile(userDataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// No user data file exists (not an error, just not logged in)
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read user data file: %w", err)
	}

	// Parse the user data
	userData := &UserData{}
	if err := json.Unmarshal(data, userData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	// Store in memory for future use
	am.userData = userData
	return userData, nil
}

// IsLoggedIn checks if the user is logged in with valid credentials
func (am *AuthManager) IsLoggedIn() bool {
	userData, err := am.LoadUserData()
	if err != nil || userData == nil {
		return false
	}

	// Check if token is expired
	if time.Now().After(userData.ExpiresAt) {
		// Token is expired, attempt to refresh it
		am.appState.Logger.Info("Token expired, attempting to refresh", "uid", userData.UID)

		// Try to refresh the token
		err := am.RefreshIDToken()
		if err != nil {
			// If refresh fails, log the user out
			am.appState.Logger.Error("Failed to refresh token", "error", err)
			am.Logout()
			return false
		}

		// Token was successfully refreshed, user is logged in
		return true
	}

	return true
}

// Logout clears the user's authentication data
func (am *AuthManager) Logout() error {
	// Clear memory
	am.userData = nil

	// Remove file
	userDataFile := filepath.Join(am.appState.DataDir, "auth", "user.json")
	if err := os.Remove(userDataFile); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove user data file: %w", err)
		}
	}

	return nil
}

// HTTPClient interface for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Default HTTP client used by the auth manager
var httpClient HTTPClient = &http.Client{Timeout: 10 * time.Second}

// RefreshIDToken refreshes the ID token using Firebase Auth REST API
func (am *AuthManager) RefreshIDToken() error {
	userData, err := am.LoadUserData()
	if err != nil || userData == nil {
		return fmt.Errorf("no user data available to refresh token")
	}

	// Get Firebase API key from environment
	apiKey := os.Getenv("FIREBASE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("Firebase API key not found in environment variables")
	}

	// Prepare the request to Firebase Auth API
	refreshURL := fmt.Sprintf("https://securetoken.googleapis.com/v1/token?key=%s", apiKey)
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": userData.RefreshToken,
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal refresh token payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", refreshURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create refresh token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send refresh token request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Read error response
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("refresh token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var refreshResponse struct {
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    string `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&refreshResponse); err != nil {
		return fmt.Errorf("failed to decode refresh token response: %w", err)
	}

	// Convert expires_in to seconds
	expiresInSeconds, err := strconv.Atoi(refreshResponse.ExpiresIn)
	if err != nil {
		// Default to 1 hour if parsing fails
		expiresInSeconds = 3600
		am.appState.Logger.Warn("Failed to parse expires_in value, using default", "error", err, "default", "1 hour")
	}

	// Update user data with new tokens
	userData.IDToken = refreshResponse.IDToken

	// Only update refresh token if a new one was provided
	if refreshResponse.RefreshToken != "" {
		userData.RefreshToken = refreshResponse.RefreshToken
	}

	// Update expiration time
	userData.ExpiresAt = time.Now().Add(time.Duration(expiresInSeconds) * time.Second)

	// Save updated user data
	if err := am.SaveUserData(userData); err != nil {
		return fmt.Errorf("failed to save refreshed user data: %w", err)
	}

	am.appState.Logger.Info("Token successfully refreshed", "uid", userData.UID)
	return nil
}

// GetCurrentUser returns the current logged-in user data
func (am *AuthManager) GetCurrentUser() (*UserData, error) {
	return am.LoadUserData()
}
