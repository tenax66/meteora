package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// struct for saving and loading preferences
type PreferencesPair struct {
	// Entry to set a preference value
	Entry *widget.Entry
	// Key for a preference value
	Key string
}

func loadPreferences(app fyne.App, pairs []PreferencesPair) {
	for _, p := range pairs {
		if value := app.Preferences().String(p.Key); value != "" {
			p.Entry.SetText(value)
		}
	}
}

func savePreferences(app fyne.App, pairs []PreferencesPair) {
	for _, p := range pairs {
		app.Preferences().SetString(p.Key, p.Entry.Text)
	}
}
