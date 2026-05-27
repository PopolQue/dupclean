package gui

import (
	"fmt"
	"log"

	"dupclean/diskanalyzer"
	"dupclean/internal/fsutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
	resultsArea   *fyne.Container
}

// DiskAnalyzerWidget creates the disk analyzer UI component
func DiskAnalyzerWidget(state *DiskAnalyzerState) fyne.CanvasObject {
	// Header
	title := canvas.NewText("Disk Analyzer", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("Identify large files and folders taking up space", theme.Color(theme.ColorNameForeground))
	subtitle.TextSize = 14
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	header := container.NewVBox(title, subtitle)

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

	// Results area
	resultsArea := container.NewStack()

	content := container.NewVBox(
		header,
		layout.NewSpacer(),
		widget.NewCard("Target Folder", "Select the directory you want to analyze", folderRow),
		container.NewHBox(layout.NewSpacer(), analyzeBtn, layout.NewSpacer()),
		container.NewVBox(progressLabel, progressBar),
		layout.NewSpacer(),
		resultsArea,
	)

	state.components = &diskAnalyzerComponents{
		analyzeBtn:    analyzeBtn,
		folderEntry:   folderEntry,
		progressLabel: progressLabel,
		progressBar:   progressBar,
		resultsArea:   resultsArea,
	}

	return container.NewScroll(content)
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
			comp.progressLabel.SetText(fmt.Sprintf("Analysis complete: %d files found", result.FileCount))
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

	// Create a list of largest directories
	list := widget.NewTable(
		func() (int, int) {
			return len(largestDirs), 2
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(theme.FolderIcon()),
				widget.NewLabel("Directory Name"),
				layout.NewSpacer(),
				widget.NewLabel("Size"),
			)
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			hbox := obj.(*fyne.Container)
			node := largestDirs[id.Row]

			label := hbox.Objects[1].(*widget.Label)
			sizeLabel := hbox.Objects[3].(*widget.Label)

			if id.Col == 0 {
				label.SetText(node.Name)
				sizeLabel.SetText(fsutil.FormatBytes(node.TotalSize))
			}
		},
	)
	list.SetColumnWidth(0, 500)
	list.SetColumnWidth(1, 150)

	// Better yet, just use a VBox with cards for the top offenders
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

	state.components.resultsArea.Objects = []fyne.CanvasObject{tabs}
	state.components.resultsArea.Refresh()
}

func NewDiskAnalyzerState(window fyne.Window) *DiskAnalyzerState {
	return &DiskAnalyzerState{
		Window: window,
	}
}
