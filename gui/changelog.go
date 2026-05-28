package gui

import (
	"fmt"

	"dupclean/internal/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

const lastSeenVersionKey = "last_seen_version"

// ShowChangelogIfNeeded checks if the current version has changed since the last run
// and shows a "What's New" popup if it has.
func ShowChangelogIfNeeded(w fyne.Window) {
	app := fyne.CurrentApp()
	lastSeen := app.Preferences().String(lastSeenVersionKey)

	// If the version has changed (or this is the first run), show the changelog
	if lastSeen != version.Version {
		showChangelog(w)
		app.Preferences().SetString(lastSeenVersionKey, version.Version)
	}
}

func showChangelog(w fyne.Window) {
	title := fmt.Sprintf("What's New in %s", version.Version)

	changelogText := `### v0.4.3 Release Highlights

**Popup Feature**
• Introduced the "What's New" popup to keep you informed about recent improvements.
• Added a manual changelog viewer in the Update screen.

**v0.4.2 Recap (Recent Improvements)**
• Fixed Cleaner CLI selection bug and UI race conditions.
• Added advanced safety triggers for permanent deletion.
• Improved hard link detection on Windows.
• Enhanced error detection for localized systems.`

	content := widget.NewRichTextFromMarkdown(changelogText)
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom(title, "Got it!", scroll, w)
	d.Show()
}
