package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DuplicateFinderWidget creates the duplicate finder UI component
func DuplicateFinderWidget(state *AppState) fyne.CanvasObject {
	// Header
	title := canvas.NewText("Duplicate Finder", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("Find and remove duplicate files by content hash", theme.Color(theme.ColorNameForeground))
	subtitle.TextSize = 14
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	header := container.NewVBox(title, subtitle)

	// Folder selection card
	folderCard := createSelectionCard(state)

	// Options card
	optionsCard := createOptionsCard(state)

	// Progress card
	progressCard := createProgressCard(state)

	// Action buttons
	scanBtn := createScanButton(state, folderCard, progressCard)

	content := container.NewVBox(
		header,
		layout.NewSpacer(),
		folderCard,
		optionsCard,
		progressCard,
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), scanBtn, layout.NewSpacer()),
		layout.NewSpacer(),
	)

	return container.NewCenter(content)
}

// DuplicateResultsWidget creates the duplicate results UI component
func DuplicateResultsWidget(state *AppState) fyne.CanvasObject {
	// Header
	title := canvas.NewText("Scan Results", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

	statsText := fmt.Sprintf(
		"%d duplicate groups | %d extra copies | %s wasted",
		len(state.Groups),
		state.Stats.TotalDupes,
		formatBytes(state.Stats.WastedBytes),
	)
	statsLabel := widget.NewLabel(statsText)
	statsLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Navigation buttons
	prevBtn := widget.NewButtonWithIcon("Previous", theme.NavigateBackIcon(), func() {
		if state.CurrentGroupIndex > 0 {
			state.CurrentGroupIndex--
			state.updateContent(DuplicateResultsWidget(state))
		}
	})
	prevBtn.Importance = widget.LowImportance

	nextBtn := widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		if state.CurrentGroupIndex < len(state.Groups)-1 {
			state.CurrentGroupIndex++
			state.updateContent(DuplicateResultsWidget(state))
		}
	})
	nextBtn.Importance = widget.LowImportance

	navButtons := container.NewHBox(prevBtn, nextBtn)

	// Group display
	groupDisplay := createGroupDisplay(state)

	// Action buttons
	skipGroupBtn := widget.NewButton("Skip Group", func() {
		state.CurrentGroupIndex++
		if state.CurrentGroupIndex >= len(state.Groups) {
			state.updateContent(createFinalUI(state))
		} else {
			state.updateContent(DuplicateResultsWidget(state))
		}
	})
	skipGroupBtn.Importance = widget.LowImportance

	skipAllBtn := widget.NewButton("Skip All", func() {
		state.updateContent(createFinalUI(state))
	})
	skipAllBtn.Importance = widget.LowImportance

	keepBtn := widget.NewButtonWithIcon("Keep #1 & Delete Others", theme.ConfirmIcon(), func() {
		if state.CurrentGroupIndex < len(state.Groups) {
			group := state.Groups[state.CurrentGroupIndex]
			keepAndDelete(state, 0, group.Files)
			if len(state.Groups) == 0 {
				state.updateContent(createFinalUI(state))
			} else {
				state.updateContent(DuplicateResultsWidget(state))
			}
		}
	})
	keepBtn.Importance = widget.HighImportance

	actionButtons := container.NewHBox(skipGroupBtn, skipAllBtn, layout.NewSpacer(), keepBtn)

	content := container.NewVBox(
		title,
		statsLabel,
		widget.NewSeparator(),
		groupDisplay,
		widget.NewSeparator(),
		actionButtons,
		widget.NewSeparator(),
		container.NewHBox(layout.NewSpacer(), navButtons, layout.NewSpacer()),
	)

	return container.NewBorder(nil, nil, nil, nil, content)
}

// DuplicateNoResultsWidget creates the "no duplicates found" UI
func DuplicateNoResultsWidget(state *AppState) fyne.CanvasObject {
	title := canvas.NewText("No Duplicates Found!", theme.Color(theme.ColorNameSuccess))
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	icon := canvas.NewImageFromResource(theme.ConfirmIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(80, 80))

	statsText := fmt.Sprintf(
		"Scanned %d files in %s\nYour files are clean!",
		state.Stats.TotalScanned,
		state.Stats.ScanDuration.Round(time.Second),
	)
	statsLabel := widget.NewLabel(statsText)
	statsLabel.Alignment = fyne.TextAlignCenter
	statsLabel.TextStyle = fyne.TextStyle{Italic: true}

	backBtn := widget.NewButtonWithIcon("Back to Home", theme.HomeIcon(), func() {
		state.updateContent(DuplicateFinderWidget(state))
	})
	backBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		icon,
		title,
		statsLabel,
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), backBtn, layout.NewSpacer()),
		layout.NewSpacer(),
	)

	return container.NewCenter(content)
}

// DuplicateFinalWidget creates the completion screen
func DuplicateFinalWidget(state *AppState) fyne.CanvasObject {
	title := canvas.NewText("Complete!", theme.Color(theme.ColorNameSuccess))
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	icon := canvas.NewImageFromResource(theme.ConfirmIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(80, 80))

	var message string
	var subMessage string
	if state.DeletedCount == 0 {
		message = "No files were deleted"
		subMessage = "Your files are safe."
	} else {
		message = fmt.Sprintf("Moved %d file(s) to Trash", state.DeletedCount)
		subMessage = fmt.Sprintf("Freed %s of disk space", formatBytes(state.FreedBytes))
	}

	resultLabel := widget.NewLabel(message)
	resultLabel.TextStyle = fyne.TextStyle{Bold: true}
	resultLabel.Alignment = fyne.TextAlignCenter

	subLabel := widget.NewLabel(subMessage)
	subLabel.TextStyle = fyne.TextStyle{Italic: true}
	subLabel.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButtonWithIcon("Start New Scan", theme.ViewRefreshIcon(), func() {
		state.Groups = nil
		state.CurrentGroupIndex = 0
		state.DeletedCount = 0
		state.FreedBytes = 0
		state.FolderPath = ""
		state.updateContent(DuplicateFinderWidget(state))
	})
	backBtn.Importance = widget.HighImportance

	quitBtn := widget.NewButtonWithIcon("Quit", theme.CancelIcon(), func() {
		state.Window.Close()
	})

	btnRow := container.NewHBox(backBtn, quitBtn)

	content := container.NewVBox(
		icon,
		title,
		resultLabel,
		subLabel,
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), btnRow, layout.NewSpacer()),
		layout.NewSpacer(),
	)

	return container.NewCenter(content)
}

// ShowDuplicateResults shows the appropriate results screen based on scan results
func ShowDuplicateResults(state *AppState) {
	if len(state.Groups) == 0 {
		state.updateContent(DuplicateNoResultsWidget(state))
		return
	}
	state.updateContent(DuplicateResultsWidget(state))
}
