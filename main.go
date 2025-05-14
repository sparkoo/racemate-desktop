package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
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
	"github.com/sparkoo/racemate-desktop/pkg/state"
	"github.com/sparkoo/racemate-desktop/pkg/upload"
)

const APP_NAME = "RaceMate"
const ACC_STATUS_LABEL_TEXT = `ACC session info: %s`
const CONTEXT_TELEMETRY = "telemetry"

func main() {
	appState, err := initApp(APP_NAME)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), state.APP_STATE, appState)

	myApp := app.New()
	myWindow := myApp.NewWindow("RaceMate")
	icon, err := fyne.LoadResourceFromPath("Icon.png")
	if err != nil {
		log.Println(fmt.Errorf("Failed to set an icon: %w", err))
	}
	myWindow.SetIcon(icon)

	// Set a fixed window size
	myWindow.Resize(fyne.NewSize(200, 300))
	myWindow.SetFixedSize(true)

	// Main UI
	label := widget.NewLabel("Hello! This is a background app.")
	myWindow.SetContent(container.NewVBox(
		label,
		widget.NewButton("Hide to Tray", func() {
			myWindow.Hide()
		}),
		widget.NewButton("Quit", func() {
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
			if appState.TelemetryOnline {
				label.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, "online"))
			} else {
				label.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, "offline"))
			}
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
		PollRate: 10 * time.Millisecond,
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
	label.SetText(text)
}

func convertToString(chars []uint16) string {
	var str string
	for _, val := range chars {
		str += string(rune(val))
	}
	return str
}
