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

	changelogText := `### v0.4.3.2 Release Highlights

**UI Improvements**

• Each update highlight now appears on its own line for better readability.

• The changelog window has been refined for better focus.

**v0.4.3.1 Recap**

• Improved the popup layout with word wrapping and vertical scrolling.

• Reduced default size for better visibility.

**v0.4.3 Features**

• Introduced the "What's New" popup to keep you informed.

• Added a manual changelog viewer in the Update screen.`

	content := widget.NewRichTextFromMarkdown(changelogText)
	content.Wrapping = fyne.TextWrapWord
	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(450, 350))

	d := dialog.NewCustom(title, "Got it!", scroll, w)
	d.Show()
}
