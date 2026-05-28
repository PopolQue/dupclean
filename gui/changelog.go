package gui

import (
	"fmt"

	"dupclean/internal/version"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
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
	// Header Component
	titleText := canvas.NewText("What's New", theme.PrimaryColor())
	titleText.TextSize = 20
	titleText.TextStyle = fyne.TextStyle{Bold: true}

	header := container.NewVBox(
		container.NewHBox(
			titleText,
			layout.NewSpacer(),
			canvas.NewText(fmt.Sprintf("%s", version.Version), theme.ForegroundColor()),
		),
		widget.NewSeparator(),
	)

	// Most Recent Update Component
	recentTitle := widget.NewLabelWithStyle("Latest Highlights", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	recentContent := widget.NewRichTextFromMarkdown(`• Improved update popup with a proper header and version info.
• Added a dedicated component for recent highlights.
• Added a structured changelog history component for better navigation.`)
	recentContent.Wrapping = fyne.TextWrapWord

	recentComponent := container.NewVBox(
		recentTitle,
		recentContent,
	)

	// Changelog History Component
	historyTitle := widget.NewLabelWithStyle("Previous Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	historyContent := widget.NewRichTextFromMarkdown(`**v0.4.3.2 Recap**
• Each update highlight now appears on its own line for better readability.
• The changelog window has been refined for better focus.

**v0.4.3.1 Recap**
• Improved the popup layout with word wrapping and vertical scrolling.
• Reduced default size for better visibility.`)
	historyContent.Wrapping = fyne.TextWrapWord

	historyComponent := container.NewVBox(
		historyTitle,
		historyContent,
	)

	// Combine components
	content := container.NewVBox(
		header,
		recentComponent,
		widget.NewSeparator(),
		historyComponent,
	)

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	d := dialog.NewCustom("DupClean Update", "Got it!", scroll, w)
	d.Show()
}
