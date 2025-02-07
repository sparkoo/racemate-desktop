package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

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
	))

	// Hide window at start
	// myWindow.Hide()

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
