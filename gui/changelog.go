package gui

import (
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
	titleLabel := widget.NewLabel("What's New")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Importance = widget.HighImportance
	titleLabel.SizeName = theme.SizeNameHeadingText

	header := container.NewVBox(
		container.NewHBox(
			titleLabel,
			layout.NewSpacer(),
			canvas.NewText(version.Version, theme.Color(theme.ColorNameForeground)),
		),
		widget.NewSeparator(),
	)

	// Most Recent Update Component
	recentTitle := widget.NewLabelWithStyle("Latest Highlights", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	recentContent := widget.NewRichTextFromMarkdown(`- **Windows Fix**: Resolved critical test failures on Windows related to system path protection.
- **Cross-Platform Robustness**: Improved platform-aware path handling in the Cache Cleaner.`)
	recentContent.Wrapping = fyne.TextWrapWord

	recentComponent := container.NewVBox(
		recentTitle,
		recentContent,
	)

	// Changelog History Component
	historyTitle := widget.NewLabelWithStyle("Previous Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	historyContent := widget.NewRichTextFromMarkdown(`**v0.4.5.0 Recap**
- UI Polish: Standardized all tool pages with a consistent header and layout.
- Windows Compatibility: Added automated Windows testing to CI.
- Code Quality: Resolved several linting warnings.

**v0.4.4.0 Recap**
- Complete UI Rebranding: Modern dark theme inspired by our new app icon.`)
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
