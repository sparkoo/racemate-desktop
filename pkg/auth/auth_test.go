package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sparkoo/racemate-desktop/pkg/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock HTTP client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Helper function to create a test AuthManager with a temporary directory
func setupTestAuthManager(t *testing.T) (*AuthManager, string, func()) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "auth-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create app state with the temporary directory
	appState := &state.AppState{
		DataDir: tempDir,
	}

	// Create auth manager
	authManager := NewAuthManager(appState)

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return authManager, tempDir, cleanup
}

// Helper to create a test user data
func createTestUserData() *UserData {
	return &UserData{
		UID:           "test-user-123",
		Email:         "test@example.com",
		DisplayName:   "Test User",
		PhotoURL:      "https://example.com/photo.jpg",
		IDToken:       "test-id-token",
		RefreshToken:  "test-refresh-token",
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		LastLoginTime: time.Now(),
	}
}

// Helper to create expired test user data
func createExpiredTestUserData() *UserData {
	return &UserData{
		UID:           "test-user-123",
		Email:         "test@example.com",
		DisplayName:   "Test User",
		PhotoURL:      "https://example.com/photo.jpg",
		IDToken:       "test-id-token",
		RefreshToken:  "test-refresh-token",
		ExpiresAt:     time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		LastLoginTime: time.Now().Add(-2 * time.Hour),
	}
}

func TestNewAuthManager(t *testing.T) {
	appState := &state.AppState{
		DataDir: "/test/dir",
	}

	authManager := NewAuthManager(appState)

	assert.NotNil(t, authManager)
	assert.Equal(t, appState, authManager.appState)
	assert.Nil(t, authManager.userData)
}

func TestSaveAndLoadUserData(t *testing.T) {
	authManager, tempDir, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create test user data
	userData := createTestUserData()

	// Save user data
	err := authManager.SaveUserData(userData)
	assert.NoError(t, err)

	// Verify file was created
	userDataFile := filepath.Join(tempDir, "auth", "user.json")
	_, err = os.Stat(userDataFile)
	assert.NoError(t, err)

	// Load user data
	loadedData, err := authManager.LoadUserData()
	assert.NoError(t, err)
	assert.NotNil(t, loadedData)

	// Verify data was loaded correctly
	assert.Equal(t, userData.UID, loadedData.UID)
	assert.Equal(t, userData.Email, loadedData.Email)
	assert.Equal(t, userData.DisplayName, loadedData.DisplayName)
	assert.Equal(t, userData.IDToken, loadedData.IDToken)
	assert.Equal(t, userData.RefreshToken, loadedData.RefreshToken)
}

func TestLoadUserDataFromMemory(t *testing.T) {
	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create and set user data in memory
	userData := createTestUserData()
	authManager.userData = userData

	// Load user data (should come from memory)
	loadedData, err := authManager.LoadUserData()
	assert.NoError(t, err)
	assert.Equal(t, userData, loadedData)
}

func TestLoadUserDataNoFile(t *testing.T) {
	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Load user data (no file exists)
	loadedData, err := authManager.LoadUserData()
	assert.NoError(t, err)
	assert.Nil(t, loadedData)
}

func TestLoadUserDataInvalidJSON(t *testing.T) {
	authManager, tempDir, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create auth directory
	authDir := filepath.Join(tempDir, "auth")
	err := os.MkdirAll(authDir, 0700)
	assert.NoError(t, err)

	// Create invalid JSON file
	userDataFile := filepath.Join(authDir, "user.json")
	err = os.WriteFile(userDataFile, []byte("invalid json"), 0600)
	assert.NoError(t, err)

	// Load user data (should fail to parse)
	loadedData, err := authManager.LoadUserData()
	assert.Error(t, err)
	assert.Nil(t, loadedData)
}

func TestIsLoggedInValid(t *testing.T) {
	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create valid user data (not expired)
	userData := createTestUserData()
	authManager.userData = userData

	// Check if logged in
	isLoggedIn := authManager.IsLoggedIn()
	assert.True(t, isLoggedIn)
}

func TestIsLoggedInNoData(t *testing.T) {
	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Check if logged in with no data
	isLoggedIn := authManager.IsLoggedIn()
	assert.False(t, isLoggedIn)
}

func TestIsLoggedInExpiredTokenRefreshFails(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &MockHTTPClient{}

	// Mock the HTTP response for a failed refresh
	mockResp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewBufferString(`{"error": "Invalid refresh token"}`)),
	}
	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// Save the original HTTP client and replace with our mock
	originalHTTPClient := httpClient
	httpClient = mockClient

	// Restore the original client after the test
	defer func() {
		httpClient = originalHTTPClient
	}()

	// Set up the environment variable for the API key
	os.Setenv("FIREBASE_API_KEY", "test-api-key")
	defer os.Unsetenv("FIREBASE_API_KEY")

	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create expired user data
	userData := createExpiredTestUserData()
	authManager.userData = userData

	// Check if logged in with expired token
	isLoggedIn := authManager.IsLoggedIn()
	assert.False(t, isLoggedIn)

	// Verify the user data was cleared (logout was called)
	assert.Nil(t, authManager.userData)
}

func TestLogout(t *testing.T) {
	authManager, tempDir, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create and save user data
	userData := createTestUserData()
	err := authManager.SaveUserData(userData)
	assert.NoError(t, err)

	// Verify file was created
	userDataFile := filepath.Join(tempDir, "auth", "user.json")
	_, err = os.Stat(userDataFile)
	assert.NoError(t, err)

	// Logout
	err = authManager.Logout()
	assert.NoError(t, err)

	// Verify memory was cleared
	assert.Nil(t, authManager.userData)

	// Verify file was removed
	_, err = os.Stat(userDataFile)
	assert.True(t, os.IsNotExist(err))
}

func TestRefreshIDTokenSuccess(t *testing.T) {
	// Create a mock HTTP client
	mockClient := &MockHTTPClient{}
	
	// Save the original HTTP client
	originalHTTPClient := httpClient

	// Create a successful refresh response
	refreshResponse := struct {
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    string `json:"expires_in"`
	}{
		IDToken:      "new-id-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    "3600",
	}

	responseBody, _ := json.Marshal(refreshResponse)
	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(responseBody)),
	}

	mockClient.On("Do", mock.Anything).Return(mockResp, nil)

	// Replace the HTTP client with our mock
	httpClient = mockClient

	// Restore the original client after the test
	defer func() {
		httpClient = originalHTTPClient
	}()

	// Set up the environment variable for the API key
	os.Setenv("FIREBASE_API_KEY", "test-api-key")
	defer os.Unsetenv("FIREBASE_API_KEY")

	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create user data
	userData := createTestUserData()
	authManager.userData = userData

	// Refresh token
	err := authManager.RefreshIDToken()
	assert.NoError(t, err)

	// Verify token was updated
	assert.Equal(t, "new-id-token", authManager.userData.IDToken)
	assert.Equal(t, "new-refresh-token", authManager.userData.RefreshToken)
}

func TestRefreshIDTokenNoAPIKey(t *testing.T) {
	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create user data
	userData := createTestUserData()
	authManager.userData = userData

	// Unset API key
	os.Unsetenv("FIREBASE_API_KEY")

	// Refresh token should fail
	err := authManager.RefreshIDToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Firebase API key not found")
}

func TestRefreshIDTokenNoUserData(t *testing.T) {
	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Refresh token with no user data
	err := authManager.RefreshIDToken()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no user data available")
}

func TestGetCurrentUser(t *testing.T) {
	authManager, _, cleanup := setupTestAuthManager(t)
	defer cleanup()

	// Create user data
	userData := createTestUserData()
	authManager.userData = userData

	// Get current user
	currentUser, err := authManager.GetCurrentUser()
	assert.NoError(t, err)
	assert.Equal(t, userData, currentUser)
}
