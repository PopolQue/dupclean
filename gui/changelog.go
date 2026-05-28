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
	recentContent := widget.NewRichTextFromMarkdown(`• **Complete UI Rebranding**: Modern dark theme inspired by our new app icon.
• **New Color Palette**: Vibrant purple and deep dark indigo tones for a premium feel.
• **Enhanced Visuals**: Added logo to the main header and branded accent bars to all sections.
• **Custom Theme**: Fully implemented custom Fyne theme for consistent branding.`)
	recentContent.Wrapping = fyne.TextWrapWord

	recentComponent := container.NewVBox(
		recentTitle,
		recentContent,
	)

	// Changelog History Component
	historyTitle := widget.NewLabelWithStyle("Previous Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	historyContent := widget.NewRichTextFromMarkdown(`**v0.4.3.4 Recap**
• Unified UI headers across all sections.
• Refined sidebar styling with better padding.
• Modernized theme API usage and resolved linting warnings.

**v0.4.3.3 Recap**
• Improved update popup with a proper header and version info.
• Added a dedicated component for recent highlights.`)
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
