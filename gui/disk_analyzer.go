package gui

import (
	"fmt"
	"log"

	"dupclean/diskanalyzer"
	"dupclean/internal/fsutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// DiskAnalyzerState holds the state for the disk analyzer widget
type DiskAnalyzerState struct {
	Window           fyne.Window
	FolderPath       string
	IsAnalyzing      bool
	Result           *diskanalyzer.AnalysisResult
	ContentContainer *fyne.Container
	components       *diskAnalyzerComponents
}

type diskAnalyzerComponents struct {
	analyzeBtn    *widget.Button
	folderEntry   *widget.Entry
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
	// Folder selection
	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("Select a folder to analyze...")
	folderEntry.SetText(state.FolderPath)

	browseBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			folderEntry.SetText(dir.Path())
			state.FolderPath = dir.Path()
		}, state.Window)
	})

	folderRow := container.NewBorder(nil, nil, nil, browseBtn, folderEntry)

	// Analyze button
	analyzeBtn := widget.NewButtonWithIcon("Analyze Disk Space", theme.StorageIcon(), func() {
		startDiskAnalysis(state)
	})
	analyzeBtn.Importance = widget.HighImportance

	// Progress
	progressLabel := widget.NewLabel("Ready")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	progressCard := widget.NewCard("", "", container.NewVBox(
		progressLabel,
		progressBar,
	))

	body := container.NewVBox(
		widget.NewCard("Target Folder", "Select the directory you want to analyze", folderRow),
		container.NewHBox(layout.NewSpacer(), analyzeBtn, layout.NewSpacer()),
		progressCard,
	)

	state.components = &diskAnalyzerComponents{
		analyzeBtn:    analyzeBtn,
		folderEntry:   folderEntry,
		progressLabel: progressLabel,
		progressBar:   progressBar,
	}

	return createToolPage("Disk Analyzer", "Identify large files and folders taking up space", body)
}

func startDiskAnalysis(state *DiskAnalyzerState) {
	if state.FolderPath == "" {
		dialog.ShowError(fmt.Errorf("please select a folder"), state.Window)
		return
	}

	state.IsAnalyzing = true
	comp := state.components
	comp.analyzeBtn.Disable()
	comp.folderEntry.Disable()
	comp.progressLabel.SetText("Analyzing filesystem...")
	comp.progressBar.Show()
	comp.progressBar.SetValue(0.5) // Indeterminate for now since Walker doesn't have granular progress

	go func() {
		opts := diskanalyzer.DefaultOptions()
		opts.MaxEntries = 500000 // Safety limit

		result, errors, err := diskanalyzer.Walk(state.FolderPath, opts)

		fyne.Do(func() {
			state.IsAnalyzing = false
			comp.analyzeBtn.Enable()
			comp.folderEntry.Enable()
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

		card := widget.NewCard(node.Name, fsutil.FormatBytes(node.TotalSize),
			widget.NewLabel(fmt.Sprintf("%d files in %s", len(node.Files), node.Path)))
		offenders.Add(card)
	}

	// Type breakdown
	typeBreakdown := diskanalyzer.TypeBreakdown(result)
	typeList := container.NewVBox()
	for i, stat := range typeBreakdown {
		if i >= 10 {
			break
		}
		row := container.NewHBox(
			widget.NewLabel(stat.Ext),
			layout.NewSpacer(),
			widget.NewLabel(fsutil.FormatBytes(stat.TotalSize)),
			widget.NewLabel(fmt.Sprintf("(%.1f%%)", stat.PctOfDisk)),
		)
		typeList.Add(row)
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

	footer := container.NewHBox(layout.NewSpacer(), backBtn, layout.NewSpacer())

	subtitle := fmt.Sprintf("Analysis complete: %d files found, %s total",
		result.FileCount, fsutil.FormatBytes(result.TotalSize))
	state.updateContent(createToolPageWithFooter("Analysis Results", subtitle, tabs, footer))
}

func NewDiskAnalyzerState(window fyne.Window) *DiskAnalyzerState {
	return &DiskAnalyzerState{
		Window: window,
	}
}
