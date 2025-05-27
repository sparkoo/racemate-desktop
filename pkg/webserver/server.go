package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"github.com/sparkoo/racemate-desktop/pkg/auth"
	"github.com/sparkoo/racemate-desktop/pkg/config"
	"github.com/sparkoo/racemate-desktop/pkg/state"
)

// Server represents the web server
type Server struct {
	server         *http.Server
	port           int
	isActive       bool
	app            fyne.App
	firebaseConfig *config.FirebaseConfig
	embeddedConfig *FirebaseConfigJSON
	timeoutTimer   *time.Timer
	timeoutDone    chan bool
	authManager    *auth.AuthManager
}

// NewServer creates a new web server instance
func NewServer(port int, app fyne.App) *Server {
	// Load Firebase config from environment variables for backward compatibility
	firebaseConfig := config.NewFirebaseConfig(
		os.Getenv("FIREBASE_API_KEY"),
		os.Getenv("FIREBASE_AUTH_DOMAIN"),
		os.Getenv("FIREBASE_PROJECT_ID"),
		os.Getenv("FIREBASE_STORAGE_BUCKET"),
		os.Getenv("FIREBASE_MESSAGING_SENDER_ID"),
		os.Getenv("FIREBASE_APP_ID"),
		os.Getenv("FIREBASE_MEASUREMENT_ID"),
	)

	// Load embedded Firebase configuration
	embeddedConfig, err := LoadEmbeddedFirebaseConfig()
	if err != nil {
		slog.Warn("Could not load embedded Firebase config, using environment variables instead", "error", err)
	}

	return &Server{
		port:           port,
		isActive:       false,
		app:            app,
		firebaseConfig: firebaseConfig,
		embeddedConfig: embeddedConfig,
	}
}

// SetAuthManager sets the auth manager for the server
func (s *Server) SetAuthManager(appState *state.AppState) {
	s.authManager = auth.NewAuthManager(appState)
}

// Start starts the web server
func (s *Server) Start() error {
	if s.isActive {
		return fmt.Errorf("server is already running")
	}

	// Initialize timeout channels
	s.timeoutDone = make(chan bool)

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/", s.handleLogin)
	mux.HandleFunc("/login", s.handleLoginSubmit)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting web server", "port", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error starting server", "error", err)
		}
	}()

	// Wait a bit to ensure server is up
	time.Sleep(100 * time.Millisecond)
	s.isActive = true

	// Set a 5-minute timeout for the server
	s.timeoutTimer = time.AfterFunc(5*time.Minute, func() {
		slog.Info("Login server timeout reached (5 minutes), stopping server")
		if err := s.Stop(); err != nil {
			slog.Error("Error stopping server on timeout", "error", err)
		}
	})

	// Open browser
	s.openBrowser()

	return nil
}

// Stop stops the web server
func (s *Server) Stop() error {
	if !s.isActive {
		return nil
	}

	// Cancel the timeout timer if it's running
	if s.timeoutTimer != nil {
		s.timeoutTimer.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	s.isActive = false
	return nil
}

// IsActive returns whether the server is active
func (s *Server) IsActive() bool {
	return s.isActive
}

// handleLogin serves the login page
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Parse the template from the embedded filesystem
	tmplContent, err := templateFS.ReadFile("templates/login.html")
	if err != nil {
		http.Error(w, "Failed to read template", http.StatusInternalServerError)
		slog.Error("Error reading template", "error", err)
		return
	}

	// Parse the template
	tmpl, err := template.New("login").Parse(string(tmplContent))
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		slog.Error("Error parsing template", "error", err)
		return
	}

	// Get Firebase configuration data for the template
	var templateData map[string]string
	
	// Prefer embedded config if available
	if s.embeddedConfig != nil {
		// Convert embedded config to template data format
		templateData = map[string]string{
			"FirebaseAPIKey":            s.embeddedConfig.APIKey,
			"FirebaseAuthDomain":        s.embeddedConfig.AuthDomain,
			"FirebaseProjectID":         s.embeddedConfig.ProjectID,
			"FirebaseStorageBucket":     s.embeddedConfig.StorageBucket,
			"FirebaseMessagingSenderID": s.embeddedConfig.MessagingSenderID,
			"FirebaseAppID":             s.embeddedConfig.AppID,
			"FirebaseMeasurementID":     s.embeddedConfig.MeasurementID,
		}
		slog.Debug("Rendering template with embedded Firebase config")
	} else {
		// Fall back to environment variable config
		templateData = s.firebaseConfig.TemplateData()
		slog.Debug("Rendering template with environment Firebase config")
	}

	// Pass Firebase configuration to the template
	err = tmpl.Execute(w, templateData)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		slog.Error("Error rendering template", "error", err)
	}
}

// handleLoginSubmit processes login form submissions with Firebase token
func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Parse JSON request body with full user data
	var userData struct {
		IDToken       string                   `json:"idToken"`
		RefreshToken  string                   `json:"refreshToken"`
		UID           string                   `json:"uid"`
		DisplayName   string                   `json:"displayName"`
		Email         string                   `json:"email"`
		PhotoURL      string                   `json:"photoURL"`
		PhoneNumber   string                   `json:"phoneNumber"`
		EmailVerified bool                     `json:"emailVerified"`
		ExpiresIn     float64                  `json:"expiresIn"` // Firebase returns this as a float
		ProviderData  []map[string]interface{} `json:"providerData"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&userData); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		slog.Error("Error decoding request body", "error", err)
		return
	}

	// TODO: Verify the Firebase ID token
	// In a production environment, you should verify the token with Firebase Admin SDK
	// For now, we'll just log the user information and create a session

	// Log user information (safely handling sensitive data)
	slog.Info("Login attempt", 
		"displayName", userData.DisplayName, 
		"uid", userData.UID, 
		"email", userData.Email)

	// Log a portion of the token (safely handling short tokens)
	tokenPreview := userData.IDToken
	if len(tokenPreview) > 20 {
		tokenPreview = tokenPreview[:20] + "..."
	}
	slog.Debug("Firebase token received", "token", tokenPreview)

	// Save user data if auth manager is available
	if s.authManager != nil {
		// Set expiration time
		expiresInSeconds := 3600 // Default to 1 hour if not provided
		if userData.ExpiresIn > 0 {
			expiresInSeconds = int(userData.ExpiresIn) // Convert float to int
		}

		// Create user data object
		authData := &auth.UserData{
			UID:          userData.UID,
			Email:        userData.Email,
			DisplayName:  userData.DisplayName,
			PhotoURL:     userData.PhotoURL,
			IDToken:      userData.IDToken,
			RefreshToken: userData.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(expiresInSeconds) * time.Second),
		}

		// Save to persistent storage
		if err := s.authManager.SaveUserData(authData); err != nil {
			slog.Error("Error saving user data", "error", err)
		} else {
			slog.Info("User data saved successfully")
		}
	} else {
		slog.Warn("Auth manager not set, user data not saved")
	}

	// Set a cookie or session to maintain login state
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "authenticated", // In production, use a secure session ID
		Path:     "/",
		MaxAge:   3600, // 1 hour
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set to true in production with HTTPS
	})

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Authentication successful",
	})

	// Schedule server shutdown after sending the response
	go func() {
		// Wait a moment to ensure the response is sent
		time.Sleep(500 * time.Millisecond)
		slog.Info("Authentication successful, stopping login server")
		// Stop the server
		if err := s.Stop(); err != nil {
			slog.Error("Error stopping server", "error", err)
		}
	}()
}

// openBrowser opens the default browser to the login page using Fyne
func (s *Server) openBrowser() {
	urlStr := fmt.Sprintf("http://localhost:%d", s.port)
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		slog.Error("Error parsing URL", "error", err)
		return
	}

	// Use Fyne's OpenURL function to open the browser
	err = s.app.OpenURL(parsedURL)
	if err != nil {
		slog.Error("Error opening browser with Fyne", "error", err)
	}
}
