package webserver

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"time"

	"fyne.io/fyne/v2"
)

// Server represents the web server
type Server struct {
	server   *http.Server
	port     int
	isActive bool
	app      fyne.App
}

// NewServer creates a new web server instance
func NewServer(port int, app fyne.App) *Server {
	return &Server{
		port:     port,
		isActive: false,
		app:      app,
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

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		log.Printf("Error rendering template: %v\n", err)
	}
}

// handleLoginSubmit processes login form submissions
func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// TODO: Implement actual authentication
	log.Printf("Login attempt: username=%s\n", username)
	
	// For now, just redirect back to the login page
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
