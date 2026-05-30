package gui

import (
	"context"
	"fmt"
	"log"

	"dupclean/diskanalyzer"
	"dupclean/gui/components"
	"dupclean/internal/fsutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DiskAnalyzerState holds the state for the disk analyzer widget
type DiskAnalyzerState struct {
	Window           fyne.Window
	ProcessManager   *ProcessManager
	FolderPath       string
	IsAnalyzing      bool
	Result           *diskanalyzer.AnalysisResult
	ContentContainer *fyne.Container
	components       *diskAnalyzerComponents
}

type diskAnalyzerComponents struct {
	analyzeBtn    *widget.Button
	progressLabel *widget.Label
	progressBar   *widget.ProgressBar
}

// updateContent updates the content container (preserves sidebar)
func (state *DiskAnalyzerState) updateContent(content fyne.CanvasObject) {
	if state.ContentContainer != nil {
		state.ContentContainer.Objects = []fyne.CanvasObject{content}
		state.ContentContainer.Refresh()
	}
}

// DiskAnalyzerWidget creates the disk analyzer UI component
func DiskAnalyzerWidget(state *DiskAnalyzerState) fyne.CanvasObject {
	// Options
	picker := components.FolderPicker("Select a folder to analyze...", state.FolderPath, false, state.Window, func(path string) {
		state.FolderPath = path
	})
	targetCard := widget.NewCard("Target Folder", "Select the directory you want to analyze", picker)

	// Scan Settings
	scanHiddenCheck := widget.NewCheck("Scan hidden files", func(b bool) {})
	followSymlinksCheck := widget.NewCheck("Follow symlinks", func(b bool) {})
	scanSettings := container.NewHBox(scanHiddenCheck, followSymlinksCheck)

	optionsCard := widget.NewCard("Scan Settings", "Configure how we identify large files", scanSettings)

	// Action
	analyzeBtn := widget.NewButtonWithIcon("Analyze Disk Space", theme.StorageIcon(), func() {
		startDiskAnalysis(state)
	})
	analyzeBtn.Importance = widget.HighImportance
	state.ProcessManager.RegisterStartButton(analyzeBtn)

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	state.components = &diskAnalyzerComponents{
		analyzeBtn:    analyzeBtn,
		progressLabel: widget.NewLabel(""),
		progressBar:   progressBar,
	}

	// Log/Menu
	logArea := container.NewVBox(state.components.progressLabel)

	return components.FixedTabLayout(
		"Disk Analyzer",
		"Identify large files and folders taking up space",
		container.NewVBox(targetCard, optionsCard),
		analyzeBtn,
		progressBar,
		logArea,
	)
}

func startDiskAnalysis(state *DiskAnalyzerState) {
	if state.FolderPath == "" {
		dialog.ShowError(fmt.Errorf("please select a folder"), state.Window)
		return
	}

	state.ProcessManager.SetProcessRunning(true)
	state.IsAnalyzing = true
	comp := state.components
	comp.analyzeBtn.Disable()
	comp.progressLabel.SetText("Analyzing filesystem...")
	comp.progressBar.Show()
	comp.progressBar.SetValue(0.5) // Indeterminate for now since Walker doesn't have granular progress

	go func() {
		opts := diskanalyzer.DefaultOptions()
		opts.MaxEntries = 500000 // Safety limit
		opts.Context = context.Background()

		result, errors, err := diskanalyzer.Walk(state.FolderPath, opts)

		fyne.Do(func() {
			state.ProcessManager.SetProcessRunning(false)
			state.IsAnalyzing = false
			comp.analyzeBtn.Enable()
			comp.progressBar.Hide()

			if err != nil {
				comp.progressLabel.SetText(fmt.Sprintf("Error: %v", err))
				dialog.ShowError(err, state.Window)
				return
			}

			if len(errors) > 0 {
				log.Printf("Analysis warnings: %d errors encountered", len(errors))
			}

			state.Result = result
			displayAnalysisResults(state)
		})
	}()
}

func displayAnalysisResults(state *DiskAnalyzerState) {
	result := state.Result
	if result == nil {
		return
	}

	// Get largest directories
	largestDirs := diskanalyzer.LargestDirs(result, 20)

	// Top offenders cards
	offenders := container.NewVBox()
	for i, node := range largestDirs {
		if i >= 10 {
			break
		}

		metadata := fmt.Sprintf("%d files", len(node.Files))
		desc := widget.NewLabel(node.Path)
		desc.Wrapping = fyne.TextWrapBreak

		card := components.ResultCard(node.Name+" ("+fsutil.FormatBytes(node.TotalSize)+")", desc, metadata, nil, nil)
		offenders.Add(card)
	}

	// Type breakdown
	typeBreakdown := diskanalyzer.TypeBreakdown(result)
	typeList := container.NewVBox()
	for i, stat := range typeBreakdown {
		if i >= 10 {
			break
		}

		metadata := fmt.Sprintf("%.1f%% of disk", stat.PctOfDisk)
		card := components.ResultCard(stat.Ext+" ("+fsutil.FormatBytes(stat.TotalSize)+")", nil, metadata, nil, nil)
		typeList.Add(card)
	}

	tabs := container.NewAppTabs(
		container.NewTabItem("Top Folders", container.NewScroll(offenders)),
		container.NewTabItem("By File Type", container.NewScroll(typeList)),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Back button
	backBtn := widget.NewButtonWithIcon("Back", theme.ContentUndoIcon(), func() {
		state.updateContent(DiskAnalyzerWidget(state))
	})

	footer := components.ActionFooter(nil, backBtn, nil)

	subtitle := fmt.Sprintf("Analysis complete: %d files found, %s total",
		result.FileCount, fsutil.FormatBytes(result.TotalSize))
	state.updateContent(components.ToolPageWithFooter("Scan Results", subtitle, tabs, footer))
}

func NewDiskAnalyzerState(window fyne.Window, pm *ProcessManager) *DiskAnalyzerState {
	return &DiskAnalyzerState{
		Window:         window,
		ProcessManager: pm,
	}
}
