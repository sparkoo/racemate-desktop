package main

import (
	"fmt"
	"time"

	"github.com/sparkoo/acctelemetry-go"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

const ACC_STATUS_LABEL_TEXT = `ACC session info: %s`

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("RaceMate")

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

	go func() {
		telemetry := acctelemetry.AccTelemetry()
		label.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, "off"))
		connected := false

		for range time.NewTicker(1 * time.Second).C {
			if connected {
				if telemetry.GraphicsPointer().ACStatus == 2 {
					fmt.Println("still conencted", telemetry.PhysicsPointer().SpeedKmh)
					label.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, "live"))
				} else {
					fmt.Println("not connected anymore")
					connected = false
					label.SetText(fmt.Sprintf(ACC_STATUS_LABEL_TEXT, "off"))
				}

			} else {
				fmt.Println("not connected, try to connect telemetry")
				if telemetry.Connect() == nil {
					fmt.Println("telemetry connected, try to get status info")
					if telemetry.GraphicsPointer().ACStatus == 2 {
						fmt.Println("status is live, connected")
						connected = true
					} else {
						fmt.Println("status is not live, disconnecting whole telemetry")
						telemetry.Close()
					}
				} else {
					fmt.Println("failed to connect telemetry, ACC probably not running at all")
				}
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
