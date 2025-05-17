package webserver

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"github.com/sparkoo/racemate-desktop/pkg/config"
)

// Server represents the web server
type Server struct {
	server         *http.Server
	port           int
	isActive       bool
	app            fyne.App
	firebaseConfig *config.FirebaseConfig
}

// NewServer creates a new web server instance
func NewServer(port int, app fyne.App) *Server {
	// Load Firebase config from environment variables
	firebaseConfig := config.NewFirebaseConfig(
		os.Getenv("FIREBASE_API_KEY"),
		os.Getenv("FIREBASE_AUTH_DOMAIN"),
		os.Getenv("FIREBASE_PROJECT_ID"),
		os.Getenv("FIREBASE_STORAGE_BUCKET"),
		os.Getenv("FIREBASE_MESSAGING_SENDER_ID"),
		os.Getenv("FIREBASE_APP_ID"),
		os.Getenv("FIREBASE_MEASUREMENT_ID"),
	)

	return &Server{
		port:           port,
		isActive:       false,
		app:            app,
		firebaseConfig: firebaseConfig,
	}
}

// Start starts the web server
func (s *Server) Start() error {
	if s.isActive {
		return fmt.Errorf("server is already running")
	}

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
		log.Printf("Starting web server on port %d\n", s.port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Error starting server: %v\n", err)
		}
	}()

	// Wait a bit to ensure server is up
	time.Sleep(100 * time.Millisecond)
	s.isActive = true

	// Open browser
	s.openBrowser()

	return nil
}

// Stop stops the web server
func (s *Server) Stop() error {
	if !s.isActive {
		return nil
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
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		log.Printf("Error loading template: %v\n", err)
		return
	}

	// Pass Firebase configuration to the template
	err = tmpl.Execute(w, s.firebaseConfig.TemplateData())
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v\n", err)
	}
}

// handleLoginSubmit processes login form submissions with Firebase token
func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Parse JSON request body
	var requestData struct {
		IDToken string `json:"idToken"`
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestData); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v\n", err)
		return
	}

	// TODO: Verify the Firebase ID token
	// In a production environment, you should verify the token with Firebase Admin SDK
	// For now, we'll just log the token and create a session

	// Log a portion of the token (safely handling short tokens)
	tokenPreview := requestData.IDToken
	if len(tokenPreview) > 20 {
		tokenPreview = tokenPreview[:20] + "..."
	}
	log.Printf("Login attempt with Firebase token: %s\n", tokenPreview)

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
		log.Println("Authentication successful, stopping login server")
		// Stop the server
		if err := s.Stop(); err != nil {
			log.Printf("Error stopping server: %v\n", err)
		}
	}()
}

// openBrowser opens the default browser to the login page using Fyne
func (s *Server) openBrowser() {
	urlStr := fmt.Sprintf("http://localhost:%d", s.port)
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		log.Printf("Error parsing URL: %v\n", err)
		return
	}

	// Use Fyne's OpenURL function to open the browser
	err = s.app.OpenURL(parsedURL)
	if err != nil {
		log.Printf("Error opening browser with Fyne: %v\n", err)
	}
}
