package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/sparkoo/racemate-desktop/pkg/acc"
	"github.com/sparkoo/racemate-desktop/pkg/constants"
)

const APP_NAME = "RaceMate"
const ACC_STATUS_LABEL_TEXT = `ACC session info: %s`
const CONTEXT_TELEMETRY = "telemetry"

func main() {
	appDataDir, err := createAppDataFolder(APP_NAME)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.WithValue(context.Background(), constants.APP_DATA_DIR_CTX_KEY, appDataDir)

	myApp := app.New()
	myWindow := myApp.NewWindow("RaceMate")
	icon, err := fyne.LoadResourceFromPath("Icon.png")
	if err != nil {
		log.Println(fmt.Errorf("Failed to set an icon: %w", err))
	}
	myWindow.SetIcon(icon)

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

	go acc.TelemetryLoop(ctx, func(ts *acc.TelemetryState) {
		if ts.Online {
			label.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, "online"))
		} else {
			label.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, "offline"))
		}
	})

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

func createAppDataFolder(appName string) (string, error) {
	var appDataDir string

	switch runtime.GOOS {
	case "windows":
		appDataDir = os.Getenv("AppData")
		if appDataDir == "" {
			appDataDir = filepath.Join(os.Getenv("LOCALAPPDATA"), appName)
		}
	default:
		return "", fmt.Errorf("Failed to create app folder in AppData: %s", appName)
	}

	appDataPath := filepath.Join(appDataDir, appName)

	if _, err := os.Stat(appDataPath); os.IsNotExist(err) {
		err := os.MkdirAll(appDataPath, 0755)
		if err != nil {
			return "", err
		}
	}
	return appDataPath, nil

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
