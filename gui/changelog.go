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

type changelogEntry struct {
	version    string
	highlights string
}

var fullChangelog = []changelogEntry{
	{
		version: "v0.4.6.1",
		highlights: `- **CI/CD Improvements**: Resolved macOS 'pkgconf' warnings and added dedicated CodeQL workflow for CGO support.
- **Better Release Notes**: Enhanced the appearance and formatting of GitHub release pages.`,
	},
	{
		version: "v0.4.6.0",
		highlights: `- **Improved Update Experience**: Update popups now only show feature highlights, excluding installation guides.
- **Dynamic Changelog**: After updating, the app now shows all release notes since your last installed version.
- **Full History**: The changelog in the About section now displays the entire project history.`,
	},
	{
		version:    "v0.4.5.4",
		highlights: `- **CI Infrastructure**: Upgraded GitHub Actions to latest major versions (v6) to natively support Node.js 24 and remove deprecation warnings.`,
	},
	{
		version: "v0.4.5.3",
		highlights: `- **CI Stability**: Resolved PowerShell syntax errors on Windows runners by enforcing bash.
- **Node.js 24**: Properly configured GitHub Actions to use Node.js 24.`,
	},
	{
		version: "v0.4.5.2",
		highlights: `- **Auto-Restart Fix**: Improved application restart reliability after updates on macOS.
- **Archive Extraction**: Enhanced robustness when extracting update binaries from archives.
- **UI Polish**: Improved "What's New" formatting for better readability.`,
	},
	{
		version: "v0.4.5.1",
		highlights: `- **Windows Fix**: Resolved critical test failures on Windows related to system path protection.
- **CI/CD**: Fixed transient network failures in GitHub Actions.`,
	},
	{
		version: "v0.4.5.0",
		highlights: `- **UI Polish**: Standardized all tool pages with a consistent header and layout.
- **Windows Compatibility**: Added automated Windows testing to CI.`,
	},
	{
		version:    "v0.4.4.0",
		highlights: `- **Complete UI Rebranding**: Modern dark theme inspired by our new app icon.`,
	},
}

// ShowChangelogIfNeeded checks if the current version has changed since the last run
// and shows a "What's New" popup if it has.
func ShowChangelogIfNeeded(w fyne.Window) {
	app := fyne.CurrentApp()
	lastSeen := app.Preferences().String(lastSeenVersionKey)

	// If the version has changed (or this is the first run), show the changelog
	if lastSeen != version.Version {
		// Only show relevant versions since last seen
		showChangelog(w, lastSeen)
		app.Preferences().SetString(lastSeenVersionKey, version.Version)
	}
}

// showChangelog shows the changelog. If sinceVersion is provided, it only shows
// entries newer than that version. If sinceVersion is empty, it shows everything.
func showChangelog(w fyne.Window, sinceVersion string) {
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

	// Build the content based on history
	content := container.NewVBox(header)

	isFirst := true
	foundLastSeen := false
	for _, entry := range fullChangelog {
		if sinceVersion != "" && entry.version == sinceVersion {
			foundLastSeen = true
			break
		}

		// Add a separator between versions
		if !isFirst {
			content.Add(widget.NewSeparator())
		}

		title := entry.version
		if isFirst && sinceVersion != "" {
			title = "Latest Highlights (" + entry.version + ")"
		}

		versionTitle := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		versionContent := widget.NewRichTextFromMarkdown(entry.highlights)
		versionContent.Wrapping = fyne.TextWrapWord

		content.Add(container.NewVBox(
			versionTitle,
			versionContent,
		))

		isFirst = false
	}

	// If we didn't find the last seen version (e.g. first run or very old),
	// just show the latest entry if we were filtering
	if sinceVersion != "" && !foundLastSeen && isFirst {
		entry := fullChangelog[0]
		versionTitle := widget.NewLabelWithStyle("Latest Highlights ("+entry.version+")", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		versionContent := widget.NewRichTextFromMarkdown(entry.highlights)
		versionContent.Wrapping = fyne.TextWrapWord
		content.Add(container.NewVBox(versionTitle, versionContent))
	}

	scroll := container.NewVScroll(content)
	scroll.SetMinSize(fyne.NewSize(500, 400))

	dialog.ShowCustom("DupClean Update", "Got it!", scroll, w)
}

// ShowFullChangelog shows the entire project history
func ShowFullChangelog(w fyne.Window) {
	showChangelog(w, "")
}
