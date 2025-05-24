package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/sparkoo/racemate-desktop/pkg/acc"
	"github.com/sparkoo/racemate-desktop/pkg/auth"
	"github.com/sparkoo/racemate-desktop/pkg/state"
	"github.com/sparkoo/racemate-desktop/pkg/upload"
	"github.com/sparkoo/racemate-desktop/pkg/webserver"
)

const APP_NAME = "RaceMate"
const ACC_STATUS_LABEL_TEXT = `ACC session info: %s`
const CONTEXT_TELEMETRY = "telemetry"
const WEB_SERVER_PORT = 12123

// Poll rate constants
const MIN_POLL_RATE = 10
const MAX_POLL_RATE = 100
const POLL_RATE_STEP = 10

func main() {
	appState, err := initApp(APP_NAME)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize auth manager
	authManager := auth.NewAuthManager(appState)

	ctx := context.WithValue(context.Background(), state.APP_STATE, appState)

	myApp := app.New()
	myWindow := myApp.NewWindow("RaceMate")
	icon, err := fyne.LoadResourceFromPath("Icon.png")
	if err != nil {
		log.Println(fmt.Errorf("Failed to set an icon: %w", err))
	}
	myWindow.SetIcon(icon)

	// Set a fixed window size
	myWindow.Resize(fyne.NewSize(250, 350))
	myWindow.SetFixedSize(true)

	// Create web server
	webServer := webserver.NewServer(WEB_SERVER_PORT, myApp)
	// Set auth manager for the web server
	webServer.SetAuthManager(appState)

	// Check if user is already logged in
	isLoggedIn := authManager.IsLoggedIn()
	userInfo := ""
	if isLoggedIn {
		user, _ := authManager.GetCurrentUser()
		if user != nil {
			userInfo = fmt.Sprintf("Logged in as: %s", user.DisplayName)
		}
	}

	// Main UI
	// Create separate labels for different information
	statusLabel := widget.NewLabel("ACC session info: offline") // Label for ACC status
	userLabel := widget.NewLabel(userInfo)                      // Label for user login info

	// Create poll rate slider and label
	// Convert time.Duration to milliseconds for the slider
	currentPollRate := int(appState.PollRate.Milliseconds())
	pollRateLabel := widget.NewLabel(fmt.Sprintf("Poll Rate: %d ms", currentPollRate))
	pollRateSlider := widget.NewSlider(float64(MIN_POLL_RATE), float64(MAX_POLL_RATE))
	pollRateSlider.Step = float64(POLL_RATE_STEP)
	pollRateSlider.Value = float64(currentPollRate)

	// Update poll rate when slider changes
	pollRateSlider.OnChanged = func(value float64) {
		// Round to nearest step
		roundedValue := int(math.Round(value/float64(POLL_RATE_STEP))) * POLL_RATE_STEP
		if roundedValue < MIN_POLL_RATE {
			roundedValue = MIN_POLL_RATE
		} else if roundedValue > MAX_POLL_RATE {
			roundedValue = MAX_POLL_RATE
		}

		// Update the slider value to show the rounded value
		pollRateSlider.Value = float64(roundedValue)

		// Update the label
		pollRateLabel.SetText(fmt.Sprintf("Poll Rate: %d ms", roundedValue))

		// Update the app state
		appState.PollRate = time.Duration(roundedValue) * time.Millisecond
		appState.Logger.Info("Poll rate updated", slog.Int("pollRate", roundedValue))
	}
	loginButton := widget.NewButton("Login", func() {
		// Start web server and open browser for login
		if webServer.IsActive() {
			userLabel.SetText("Login server is already running")
			return
		}

		err := webServer.Start()
		if err != nil {
			userLabel.SetText(fmt.Sprintf("Error starting login server: %v", err))
			log.Printf("Error starting login server: %v\n", err)
			return
		}

		userLabel.SetText("Login server started. Browser should open automatically.")
	})

	logoutButton := widget.NewButton("Logout", func() {
		err := authManager.Logout()
		if err != nil {
			userLabel.SetText(fmt.Sprintf("Error logging out: %v", err))
			log.Printf("Error logging out: %v\n", err)
			return
		}
		userLabel.SetText("Logged out successfully")
	})

	// Show login or logout button based on current state
	var authButtons *fyne.Container
	if isLoggedIn {
		authButtons = container.NewVBox(userLabel, logoutButton)
	} else {
		authButtons = container.NewVBox(loginButton)
	}

	myWindow.SetContent(container.NewVBox(
		statusLabel, // ACC status label
		authButtons,
		// Poll rate controls
		container.NewVBox(
			pollRateLabel,
			pollRateSlider,
		),
		widget.NewButton("Hide to Tray", func() {
			myWindow.Hide()
		}),
		widget.NewButton("Quit", func() {
			// Stop web server if running
			if webServer.IsActive() {
				if err := webServer.Stop(); err != nil {
					log.Printf("Error stopping web server: %v\n", err)
				}
			}
			myApp.Quit()
		}),
	))

	// Hide window at start
	// myWindow.Hide()

	go acc.TelemetryLoop(ctx)

	go upload.UploadJob(ctx)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			// Thread-safe UI update using fyne.Do()
			status := "offline"
			if appState.TelemetryOnline {
				status = "online"
			}
			final := status // Create a local copy for the closure

			// Check login status periodically
			currentlyLoggedIn := authManager.IsLoggedIn()
			var userDisplayInfo string
			if currentlyLoggedIn {
				user, _ := authManager.GetCurrentUser()
				if user != nil {
					userDisplayInfo = fmt.Sprintf("Logged in as: %s", user.DisplayName)
				}
			}

			// Use fyne.Do to safely update UI from a goroutine
			fyne.Do(func() {
				// Update ACC status label
				statusLabel.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, final))

				// Update login status if it changed
				if currentlyLoggedIn != isLoggedIn {
					isLoggedIn = currentlyLoggedIn

					// Rebuild auth buttons based on new state
					if isLoggedIn {
						userLabel.SetText(userDisplayInfo)
						authButtons.Objects = []fyne.CanvasObject{userLabel, logoutButton}
					} else {
						userLabel.SetText("")
						authButtons.Objects = []fyne.CanvasObject{loginButton}
					}
					authButtons.Refresh()
				}
			})
		}
	}()

	// System Tray Support
	if deskApp, ok := myApp.(desktop.App); ok {
		menu := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show Window", func() {
				myWindow.Show()
			}),
			fyne.NewMenuItem("Quit", func() {
				myApp.Quit()
			}),
		)
		deskApp.SetSystemTrayMenu(menu)
		deskApp.SetSystemTrayIcon(icon)
	}

	myWindow.ShowAndRun()
}

func initApp(appName string) (*state.AppState, error) {
	appState := &state.AppState{
		PollRate:  10 * time.Millisecond,
		UploadURL: "https://lapupload-hwppiybqxq-ey.a.run.app",
	}

	if err := initDataDirs(appName, appState); err != nil {
		return nil, fmt.Errorf("failed to init data dirs: %w", err)
	}

	initLogger(appState)

	return appState, nil
}

func initLogger(appState *state.AppState) {
	appState.Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func initDataDirs(appName string, appState *state.AppState) error {
	var appDataDir string

	switch runtime.GOOS {
	case "windows":
		appDataDir = os.Getenv("AppData")
		if appDataDir == "" {
			appDataDir = filepath.Join(os.Getenv("LOCALAPPDATA"), appName)
		}
	default:
		return fmt.Errorf("We can do only Windows: %s", appName)
	}

	appDir := filepath.Join(appDataDir, appName)
	if err := createFullDir(appDir); err != nil {
		return fmt.Errorf("failed to create an app dir '%s': %w", appDir, err)
	} else {
		appState.DataDir = appDir
	}

	uploadDir := filepath.Join(appDataDir, appName, "upload")
	if err := createFullDir(uploadDir); err != nil {
		return fmt.Errorf("failed to create an upload dir '%s': %w", uploadDir, err)
	} else {
		appState.UploadDir = uploadDir
	}

	uploadedDir := filepath.Join(appDataDir, appName, "uploaded")
	if err := createFullDir(uploadedDir); err != nil {
		return fmt.Errorf("failed to create an uploaded dir '%s': %w", uploadedDir, err)
	} else {
		appState.UploadedDir = uploadedDir
	}

	logsDir := filepath.Join(appDataDir, appName, "logs")
	if err := createFullDir(logsDir); err != nil {
		return fmt.Errorf("failed to create a logs dir '%s': %w", logsDir, err)
	} else {
		appState.LogsDir = logsDir
	}

	return nil
}

func createFullDir(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateLabel(label *widget.Label, text string) {
	// Use fyne.Do to ensure thread-safe UI updates
	fyne.Do(func() {
		label.SetText(text)
	})
}

func convertToString(chars []uint16) string {
	var str string
	for _, val := range chars {
		str += string(rune(val))
	}
	return str
}
