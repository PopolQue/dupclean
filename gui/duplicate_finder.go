package gui

import (
	"fmt"
	"time"

	"dupclean/internal/fsutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DuplicateFinderWidget creates the duplicate finder UI component
func DuplicateFinderWidget(state *AppState) fyne.CanvasObject {
	// Folder selection card
	folderCard := createSelectionCard(state)

	// Options card
	optionsCard := createOptionsCard(state)

	// Progress card
	progressCard := createProgressCard(state)

	// Action buttons
	scanBtn := createScanButton(state, folderCard, progressCard)

	body := container.NewVBox(
		folderCard,
		optionsCard,
		progressCard,
		container.NewHBox(layout.NewSpacer(), scanBtn, layout.NewSpacer()),
	)

	return createToolPage("Duplicate Finder", "Find and remove duplicate files by content hash", body)
}

// DuplicateResultsWidget creates the duplicate results UI component
func DuplicateResultsWidget(state *AppState) fyne.CanvasObject {
	statsText := fmt.Sprintf(
		"%d groups | %d extra copies | %s wasted",
		len(state.Groups),
		state.Stats.TotalDupes,
		fsutil.FormatBytes(state.Stats.WastedBytes),
	)

	// Group display
	groupDisplay := createGroupDisplay(state)

	// Action buttons
	cancelBtn := widget.NewButton("Cancel", func() {
		state.updateContent(DuplicateFinderWidget(state))
	})
	cancelBtn.Importance = widget.LowImportance

	smartBtn := widget.NewButton("Smart Select", func() {
		state.mu.Lock()
		for i := range state.Selections {
			for j := range state.Selections[i] {
				state.Selections[i][j] = (j == 0) // Keep first one
			}
		}
		state.mu.Unlock()
		state.updateContent(DuplicateResultsWidget(state))
	})

	cleanBtn := widget.NewButtonWithIcon("Clean Selected", theme.DeleteIcon(), func() {
		dialog.ShowConfirm(
			"Clean Selected Files?",
			"This will move all unselected files to the Trash. Are you sure?",
			func(ok bool) {
				if ok {
					cleanSelected(state)
				}
			},
			state.Window,
		)
	})
	cleanBtn.Importance = widget.HighImportance

	actionButtons := container.NewHBox(cancelBtn, layout.NewSpacer(), smartBtn, cleanBtn)

	return createToolPageWithFooter("Scan Results", statsText, groupDisplay, actionButtons)
}

// DuplicateNoResultsWidget creates the "no duplicates found" UI
func DuplicateNoResultsWidget(state *AppState) fyne.CanvasObject {
	statsText := fmt.Sprintf(
		"Scanned %d files in %s. Your files are clean!",
		state.Stats.TotalScanned,
		state.Stats.ScanDuration.Round(time.Second),
	)

	backBtn := widget.NewButtonWithIcon("Back to Home", theme.HomeIcon(), func() {
		state.updateContent(DuplicateFinderWidget(state))
	})
	backBtn.Importance = widget.HighImportance

	return createStatusPage(
		"Scan Complete",
		"No duplicates were found",
		theme.ConfirmIcon(),
		"No Duplicates Found!",
		statsText,
		backBtn,
	)
}

// DuplicateFinalWidget creates the completion screen
func DuplicateFinalWidget(state *AppState) fyne.CanvasObject {
	var subMessage string
	if state.DeletedCount == 0 {
		subMessage = "No files were deleted. Your files are safe."
	} else {
		subMessage = fmt.Sprintf("Moved %d file(s) to Trash\nFreed %s of disk space", state.DeletedCount, fsutil.FormatBytes(state.FreedBytes))
	}

	if state.SkippedCount > 0 {
		subMessage += fmt.Sprintf("\n⚠️ %d file(s) skipped (modified since scan)", state.SkippedCount)
	}

	backBtn := widget.NewButtonWithIcon("Start New Scan", theme.ViewRefreshIcon(), func() {
		state.Groups = nil
		state.CurrentGroupIndex = 0
		state.DeletedCount = 0
		state.FreedBytes = 0
		state.SkippedCount = 0
		state.SkippedFiles = nil
		state.FolderPath = ""
		state.updateContent(DuplicateFinderWidget(state))
	})
	backBtn.Importance = widget.HighImportance

	quitBtn := widget.NewButtonWithIcon("Quit", theme.CancelIcon(), func() {
		state.Window.Close()
	})

	btnRow := container.NewHBox(backBtn, quitBtn)

	return createStatusPage(
		"Cleaning Finished",
		"Summary of the cleaning operation",
		theme.ConfirmIcon(),
		"Complete!",
		subMessage,
		btnRow,
	)
}

// ShowDuplicateResults shows the appropriate results screen based on scan results
func ShowDuplicateResults(state *AppState) {
	if len(state.Groups) == 0 {
		state.updateContent(DuplicateNoResultsWidget(state))
		return
	}
	state.updateContent(DuplicateResultsWidget(state))
}
