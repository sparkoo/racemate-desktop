package main

import (
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/sparkoo/racemate-desktop/pkg/acc"
)

const ACC_STATUS_LABEL_TEXT = `ACC session info: %s`
const CONTEXT_TELEMETRY = "telemetry"

func main() {
	fmt.Println(os.Getenv("LOCALAPPDATA"))

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

	go acc.TelemetryLoop(func(ts *acc.TelemetryState) {
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
