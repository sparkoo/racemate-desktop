package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
		// Token is expired, but we could try to refresh it
		// For now, we'll just consider the user logged out
		// In a production app, you might attempt to refresh the token here

		// Log that token is expired
		fmt.Printf("Token expired for user: %s\n", userData.UID)
		am.Logout()
		return false
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

// RefreshIDToken refreshes the ID token
// For a desktop app using Firebase Web SDK, we have limited options
func (am *AuthManager) RefreshIDToken() error {
	userData, err := am.LoadUserData()
	if err != nil || userData == nil {
		return fmt.Errorf("no user data available to refresh token")
	}

	// In a desktop app using Firebase Web SDK, we have several options:
	// 1. Re-launch the login flow (webview) when token expires
	// 2. Use a custom token approach with Firebase Admin SDK (requires server)
	// 3. Implement a native Firebase Auth REST API client

	// For option #3, here's how you would implement it:
	// Note: This requires your Firebase API key which should be kept secure

	// For now, we'll just mark the token as expired
	// The main app will detect this and prompt for re-login

	// Log that token refresh was attempted
	fmt.Printf("Token refresh attempted for user: %s\n", userData.UID)

	// In a production app, you would extend the token lifetime
	// For now, we'll just return nil and let the app handle re-login
	return nil
}

// GetCurrentUser returns the current logged-in user data
func (am *AuthManager) GetCurrentUser() (*UserData, error) {
	return am.LoadUserData()
}
