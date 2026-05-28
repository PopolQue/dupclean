package gui

import (
	"fmt"
	"time"

	"dupclean/internal/fsutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
		layout.NewSpacer(),
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

	body := container.NewVBox(
		groupDisplay,
		widget.NewSeparator(),
		actionButtons,
	)

	return createToolPage("Scan Results", statsText, body)
}

// DuplicateNoResultsWidget creates the "no duplicates found" UI
func DuplicateNoResultsWidget(state *AppState) fyne.CanvasObject {
	title := canvas.NewText("No Duplicates Found!", theme.Color(theme.ColorNameSuccess))
	title.TextSize = 28
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

	body := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(container.NewVBox(
			icon,
			title,
			statsLabel,
			layout.NewSpacer(),
			container.NewHBox(layout.NewSpacer(), backBtn, layout.NewSpacer()),
		)),
		layout.NewSpacer(),
	)

	return createToolPage("Scan Complete", "No duplicates were found", body)
}

// DuplicateFinalWidget creates the completion screen
func DuplicateFinalWidget(state *AppState) fyne.CanvasObject {
	title := canvas.NewText("Complete!", theme.Color(theme.ColorNameSuccess))
	title.TextSize = 28
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
		subMessage = fmt.Sprintf("Freed %s of disk space", fsutil.FormatBytes(state.FreedBytes))
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

	body := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(container.NewVBox(
			icon,
			title,
			resultLabel,
			subLabel,
			layout.NewSpacer(),
			container.NewHBox(layout.NewSpacer(), btnRow, layout.NewSpacer()),
		)),
		layout.NewSpacer(),
	)

	return createToolPage("Cleaning Finished", "Summary of the cleaning operation", body)
}

// ShowDuplicateResults shows the appropriate results screen based on scan results
func ShowDuplicateResults(state *AppState) {
	if len(state.Groups) == 0 {
		state.updateContent(DuplicateNoResultsWidget(state))
		return
	}
	state.updateContent(DuplicateResultsWidget(state))
}
