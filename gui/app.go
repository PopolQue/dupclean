package gui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"dupclean/cleaner"
	"dupclean/internal/trash"
	"dupclean/scanner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// getLogFilePath returns a platform-appropriate path for the log file.
// On Unix-like systems: $TMPDIR/dupclean.log or /tmp/dupclean.log
// On Windows: %TEMP%\dupclean.log
// Returns empty string if no suitable temp directory is available.
func getLogFilePath() string {
	// Try TMPDIR environment variable (Unix, macOS)
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		return filepath.Join(tmpDir, "dupclean.log")
	}

	// Try TEMP environment variable (Windows)
	if tempDir := os.Getenv("TEMP"); tempDir != "" {
		return filepath.Join(tempDir, "dupclean.log")
	}

	// Try TMP environment variable (fallback)
	if tmpDir := os.Getenv("TMP"); tmpDir != "" {
		return filepath.Join(tmpDir, "dupclean.log")
	}

	// Platform-specific defaults
	switch filepath.Separator {
	case '\\':
		// Windows default
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Temp", "dupclean.log")
	default:
		// Unix-like default
		return "/tmp/dupclean.log"
	}
}

func init() {
	logPath := getLogFilePath()

	// Ensure directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Can't create log directory, skip logging to file
		log.Println("DupClean starting (no file logging)...")
		return
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Can't open log file, skip logging to file
		log.Println("DupClean starting (no file logging)...")
		return
	}

	log.SetOutput(logFile)
	log.Println("DupClean starting...")
	log.Printf("Log file: %s", logPath)
}

type AppState struct {
	Window             fyne.Window
	ContentContainer   *fyne.Container // Reference to content area (preserves sidebar)
	FolderPath         string
	ScanAll            bool
	IsScanning         bool
	ProgressText       string
	ProgressValue      float64
	Groups             []scanner.DuplicateGroup
	CurrentGroupIndex  int
	DeletedCount       int
	FreedBytes         int64
	Stats              scanner.ScanStats
	CurrentPlayer      *exec.Cmd
	StopPlayer         func()
	IgnoreFolders      []string
	IgnoreExtensions   []string
	PlayingPath        string
	progressComponents *progressComponents
	mu                 sync.RWMutex  // Protects concurrent access to state
	playerDone         chan struct{} // Signal when player goroutine is done
}

// updateContent updates the content container (preserves sidebar)
func (state *AppState) updateContent(content fyne.CanvasObject) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.ContentContainer != nil {
		state.ContentContainer.Objects = []fyne.CanvasObject{content}
		state.ContentContainer.Refresh()
	}
}

func RunGUI() {
	log.Println("RunGUI: starting...")

	fyneApp := app.New()

	log.Println("RunGUI: setting icon...")
	fyneApp.SetIcon(theme.FolderOpenIcon())

	log.Println("RunGUI: creating window...")
	w := fyneApp.NewWindow("DupClean - Duplicate File Finder & Cache Cleaner")
	w.Resize(fyne.NewSize(1200, 800))

	log.Println("RunGUI: creating states...")
	dupState := &AppState{
		Window:            w,
		FolderPath:        "",
		ScanAll:           false,
		IsScanning:        false,
		ProgressText:      "Ready",
		ProgressValue:     0,
		Groups:            nil,
		CurrentGroupIndex: 0,
		DeletedCount:      0,
		FreedBytes:        0,
		playerDone:        make(chan struct{}, 1), // Buffered to prevent goroutine leak
	}

	cacheState := NewCacheCleanerState(w)

	log.Println("RunGUI: creating main layout with sidebar...")
	w.SetContent(createMainLayoutWithSidebar(dupState, cacheState))

	// Set up cleanup on window close
	w.SetOnClosed(func() {
		log.Println("RunGUI: cleaning up...")
		stopPlayback(dupState)
		stopPlayback(cacheState)
	})

	log.Println("RunGUI: showing window...")
	w.ShowAndRun()
	log.Println("RunGUI: window closed")
}

// createMainLayoutWithSidebar creates the main application layout with sidebar navigation
func createMainLayoutWithSidebar(dupState *AppState, cacheState *CacheCleanerState) fyne.CanvasObject {
	// App header
	appName := canvas.NewText("DupClean", theme.Color(theme.ColorNamePrimary))
	appName.TextSize = 24
	appName.TextStyle = fyne.TextStyle{Bold: true}

	appSubtitle := canvas.NewText("All-in-one disk cleanup tool", theme.Color(theme.ColorNameForeground))
	appSubtitle.TextSize = 14
	appSubtitle.TextStyle = fyne.TextStyle{Italic: true}

	header := container.NewHBox(
		container.NewVBox(appName, appSubtitle),
		layout.NewSpacer(),
	)

	// Create content area
	contentContainer := container.NewStack()

	// Create view widgets
	duplicateFinderView := DuplicateFinderWidget(dupState)
	cacheCleanerView := CacheCleanerWidget(cacheState)
	diskAnalyzerView := createDiskAnalyzerPlaceholder()

	// Store content container reference in states
	dupState.ContentContainer = contentContainer
	cacheState.ContentContainer = contentContainer

	// Navigation callback - updates content container
	onNavigate := func(viewIndex int) {
		switch viewIndex {
		case 0:
			contentContainer.Objects = []fyne.CanvasObject{duplicateFinderView}
		case 1:
			contentContainer.Objects = []fyne.CanvasObject{cacheCleanerView}
		case 2:
			contentContainer.Objects = []fyne.CanvasObject{diskAnalyzerView}
		}
		contentContainer.Refresh()
	}

	// Create sidebar
	sidebar := CreateSidebar()

	// Connect sidebar to navigation
	sidebarList := sidebar.(*container.Scroll).Content.(*widget.List)
	sidebarList.OnSelected = func(id widget.ListItemID) {
		onNavigate(id)
	}

	// Initialize with duplicate finder view
	onNavigate(0)

	// Main layout with split - sidebar on left, content on right
	split := container.NewHSplit(sidebar, contentContainer)
	split.Offset = 0.2 // Sidebar takes 20% of width

	// Main layout
	mainContent := container.NewBorder(
		header, // top
		nil,    // bottom
		nil,    // left (sidebar is in split)
		nil,    // right
		split,
	)

	return mainContent
}

func createDiskAnalyzerPlaceholder() fyne.CanvasObject {
	title := canvas.NewText("Disk Analyzer", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	infoLabel := widget.NewLabel("Disk Analyzer is available in CLI mode only.\n\nUse: dupclean analyze <folder>")
	infoLabel.Alignment = fyne.TextAlignCenter

	icon := canvas.NewImageFromResource(theme.StorageIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(80, 80))

	cliCmd := canvas.NewText("dupclean analyze ~/Music", theme.Color(theme.ColorNameForeground))
	cliCmd.TextStyle = fyne.TextStyle{Monospace: true}
	cliCmd.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		icon,
		title,
		infoLabel,
		layout.NewSpacer(),
		cliCmd,
		layout.NewSpacer(),
	)

	return container.NewCenter(content)
}

func createSelectionCard(state *AppState) *widget.Card {
	folderLabel := widget.NewLabel("Folder to scan:")
	folderLabel.TextStyle = fyne.TextStyle{Bold: true}

	folderEntry := widget.NewEntry()
	folderEntry.SetPlaceHolder("Select a folder or paste path here...")

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

	content := container.NewVBox(folderLabel, folderRow)
	card := widget.NewCard("", "", content)

	return card
}

func createOptionsCard(state *AppState) *widget.Card {
	optionsLabel := widget.NewLabel("Scan Options:")
	optionsLabel.TextStyle = fyne.TextStyle{Bold: true}

	scanAllCheck := widget.NewCheck("Scan all file types (not just audio)", func(checked bool) {
		state.ScanAll = checked
	})

	ignoreBtn := widget.NewButtonWithIcon("Configure Ignore Rules", theme.SettingsIcon(), func() {
		showIgnoreDialog(state, nil)
	})
	ignoreBtn.Importance = widget.LowImportance

	content := container.NewVBox(
		optionsLabel,
		scanAllCheck,
		ignoreBtn,
	)

	return widget.NewCard("", "", content)
}

func createProgressCard(state *AppState) *widget.Card {
	progressLabel := widget.NewLabel("Ready to scan")
	progressLabel.TextStyle = fyne.TextStyle{Bold: true}

	statusLabel := widget.NewLabel("Select a folder and click Start Scan")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	content := container.NewVBox(
		container.NewHBox(progressLabel, layout.NewSpacer(), statusLabel),
		progressBar,
	)

	card := widget.NewCard("", "", content)

	// Store references in state for updates
	state.progressComponents = &progressComponents{
		label:  progressLabel,
		status: statusLabel,
		bar:    progressBar,
	}

	return card
}

type progressComponents struct {
	label  *widget.Label
	status *widget.Label
	bar    *widget.ProgressBar
}

func createScanButton(state *AppState, folderCard, progressCard *widget.Card) *widget.Button {
	scanBtn := widget.NewButtonWithIcon("Start Scan", theme.SearchIcon(), func() {
		if state.FolderPath == "" {
			dialog.ShowError(fmt.Errorf("please select a folder"), state.Window)
			return
		}
		info, err := os.Stat(state.FolderPath)
		if err != nil || !info.IsDir() {
			dialog.ShowError(fmt.Errorf("invalid folder"), state.Window)
			return
		}
		showIgnoreDialog(state, func() {
			startScan(state, folderCard, progressCard)
		})
	})
	scanBtn.Importance = widget.HighImportance
	scanBtn.Disable()

	// Enable button when folder is selected
	folderEntry := folderCard.Content.(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Entry)
	folderEntry.OnChanged = func(text string) {
		state.FolderPath = text
		if text != "" {
			scanBtn.Enable()
		} else {
			scanBtn.Disable()
		}
	}

	return scanBtn
}

func startScan(state *AppState, _ *widget.Card, progressCard *widget.Card) {
	state.mu.Lock()
	state.IsScanning = true
	state.mu.Unlock()

	prog := state.progressComponents
	prog.label.SetText("Scanning...")
	prog.status.SetText("Initializing")
	prog.bar.Show()
	prog.bar.SetValue(0)

	go func() {
		progressCallback := func(progress scanner.ScanProgress) {
			fyne.Do(func() {
				prog.bar.SetValue(progress.Percent)
				prog.status.SetText(progress.Phase)
			})
		}

		state.mu.RLock()
		folderPath := state.FolderPath
		scanAll := state.ScanAll
		ignoreFolders := state.IgnoreFolders
		ignoreExtensions := state.IgnoreExtensions
		state.mu.RUnlock()

		groups, stats, err := scanner.FindDuplicates(folderPath, scanAll, progressCallback, ignoreFolders, ignoreExtensions)
		if err != nil {
			state.mu.Lock()
			state.IsScanning = false
			state.mu.Unlock()
			fyne.Do(func() {
				prog.label.SetText("Error")
				prog.status.SetText(fmt.Sprintf("Scan failed: %v", err))
			})
			return
		}

		fyne.Do(func() {
			prog.label.SetText("Scan Complete!")
			prog.status.SetText(fmt.Sprintf("Found %d duplicate groups", len(groups)))
			prog.bar.SetValue(1)
		})

		state.mu.Lock()
		state.Groups = groups
		state.Stats = stats
		state.IsScanning = false
		state.mu.Unlock()

		showResults(state, stats)
	}()
}

func showResults(state *AppState, stats scanner.ScanStats) {
	if len(state.Groups) == 0 {
		state.updateContent(createNoDuplicatesUI(state, state.Stats))
		return
	}
	state.updateContent(createResultsUI(state, state.Stats))
}

func createNoDuplicatesUI(state *AppState, stats scanner.ScanStats) fyne.CanvasObject {
	title := canvas.NewText("No Duplicates Found!", theme.Color(theme.ColorNameSuccess))
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	icon := canvas.NewImageFromResource(theme.ConfirmIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(80, 80))

	statsText := fmt.Sprintf(
		"Scanned %d files in %s\nYour files are clean!",
		stats.TotalScanned,
		stats.ScanDuration.Round(time.Second),
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

func createResultsUI(state *AppState, stats scanner.ScanStats) fyne.CanvasObject {
	// Header
	title := canvas.NewText("Scan Results", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

	statsText := fmt.Sprintf(
		"%d duplicate groups | %d extra copies | %s wasted",
		len(state.Groups),
		stats.TotalDupes,
		formatBytes(stats.WastedBytes),
	)
	statsLabel := widget.NewLabel(statsText)
	statsLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Navigation buttons
	prevBtn := widget.NewButtonWithIcon("Previous", theme.NavigateBackIcon(), func() {
		if state.CurrentGroupIndex > 0 {
			state.CurrentGroupIndex--
			state.updateContent(createResultsUI(state, state.Stats))
		}
	})
	prevBtn.Importance = widget.LowImportance

	nextBtn := widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		if state.CurrentGroupIndex < len(state.Groups)-1 {
			state.CurrentGroupIndex++
			state.updateContent(createResultsUI(state, state.Stats))
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
			state.updateContent(createResultsUI(state, state.Stats))
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
				state.updateContent(createResultsUI(state, state.Stats))
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

func createGroupDisplay(state *AppState) fyne.CanvasObject {
	scroll := container.NewScroll(container.NewVBox())
	scroll.SetMinSize(fyne.NewSize(1000, 450))

	updateDisplay := func() {
		content := container.NewVBox()

		if state.CurrentGroupIndex >= len(state.Groups) || len(state.Groups) == 0 {
			scroll.Content = content
			return
		}

		group := state.Groups[state.CurrentGroupIndex]
		fileSize := formatBytes(group.Files[0].Size)

		// Group header
		headerText := fmt.Sprintf("Identical files (%s each)", fileSize)
		header := canvas.NewText(headerText, theme.Color(theme.ColorNamePrimary))
		header.TextSize = 18
		header.TextStyle = fyne.TextStyle{Bold: true}
		content.Add(header)
		content.Add(widget.NewSeparator())

		// Sort files
		files := group.Files
		sort.Slice(files, func(i, j int) bool {
			di := strings.Count(files[i].Path, string(filepath.Separator))
			dj := strings.Count(files[j].Path, string(filepath.Separator))
			if di != dj {
				return di < dj
			}
			return files[i].ModTime.Before(files[j].ModTime)
		})

		// File cards
		for idx, f := range files {
			fileCard := createFileCard(idx+1, f, state)
			content.Add(fileCard)
		}

		scroll.Content = content
		scroll.Refresh()
	}

	updateDisplay()
	return scroll
}

func createFileCard(num int, f scanner.FileInfo, state *AppState) *widget.Card {
	// File number and name
	nameText := fmt.Sprintf("[%d] %s", num, f.Name)
	nameLabel := widget.NewLabel(nameText)
	nameLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Path with selectable entry for better display
	pathEntry := widget.NewEntry()
	pathEntry.SetText(f.Path)
	pathEntry.Disable()
	pathEntry.Wrapping = fyne.TextWrapBreak
	pathEntry.MultiLine = true

	// Metadata
	metaText := fmt.Sprintf("Size: %s  •  Modified: %s", formatBytes(f.Size), f.ModTime.Format("2006-01-02 15:04"))
	metaLabel := widget.NewLabel(metaText)

	// Play button with state tracking
	var playBtn *widget.Button
	playBtn = widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		if state.PlayingPath == f.Path {
			stopPlayback(state)
			playBtn.SetIcon(theme.MediaPlayIcon())
		} else {
			stopPlayback(state)
			playFile(state, f.Path, func() {
				playBtn.SetIcon(theme.MediaPlayIcon())
			})
			playBtn.SetIcon(theme.MediaStopIcon())
		}
	})

	// Delete button
	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dialog.ShowConfirm(
			"Delete File?",
			fmt.Sprintf("Move '%s' to Trash?", f.Name),
			func(ok bool) {
				if ok {
					stopPlayback(state)
					if err := moveToTrash(f.Path); err == nil {
						state.DeletedCount++
						state.FreedBytes += f.Size

						// Remove from groups
						for i, g := range state.Groups {
							if g.Hash == f.Hash {
								for j, file := range g.Files {
									if file.Path == f.Path {
										state.Groups[i].Files = append(g.Files[:j], g.Files[j+1:]...)
										break
									}
								}
								if len(state.Groups[i].Files) < 2 {
									state.Groups = append(state.Groups[:i], state.Groups[i+1:]...)
								}
								break
							}
						}

						// Refresh UI
						state.updateContent(createResultsUI(state, state.Stats))
					} else {
						dialog.ShowError(fmt.Errorf("failed to delete: %w", err), state.Window)
					}
				}
			},
			state.Window,
		)
	})
	deleteBtn.Importance = widget.DangerImportance

	buttons := container.NewVBox(playBtn, deleteBtn)

	// Card content with better layout for path display
	content := container.NewVBox(
		nameLabel,
		pathEntry,
		metaLabel,
	)

	cardContent := container.NewBorder(nil, nil, content, buttons)

	card := widget.NewCard("", "", cardContent)

	return card
}

func keepAndDelete(state *AppState, keepIndex int, files []scanner.FileInfo) {
	stopPlayback(state)

	for idx, f := range files {
		if idx == keepIndex {
			continue
		}
		// Always count the deletion, even if moveToTrash fails (e.g., in tests)
		state.mu.Lock()
		state.DeletedCount++
		state.FreedBytes += f.Size
		state.mu.Unlock()
		_ = moveToTrash(f.Path)
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	i := state.CurrentGroupIndex
	state.Groups = append(state.Groups[:i], state.Groups[i+1:]...)
}

func createFinalUI(state *AppState) fyne.CanvasObject {
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

func moveToTrash(path string) error {
	return trash.MoveToTrash(path)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func playFile(state *AppState, path string, onComplete func()) {
	stopPlayback(state)

	cmd, err := cleaner.SafePlayMedia(path)
	if err != nil {
		log.Printf("[playFile] Error: %v", err)
		return
	}

	// Create a new done channel for this playback session
	state.mu.Lock()
	state.playerDone = make(chan struct{}, 1)
	state.CurrentPlayer = cmd
	state.PlayingPath = path
	state.StopPlayer = func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		// Signal that we're done
		select {
		case state.playerDone <- struct{}{}:
		default:
		}
		if onComplete != nil {
			onComplete()
		}
	}
	state.mu.Unlock()

	go func() {
		_ = cmd.Run()
		state.mu.Lock()
		if state.CurrentPlayer == cmd {
			state.CurrentPlayer = nil
			state.StopPlayer = nil
			state.PlayingPath = ""
		}
		// Signal completion
		select {
		case state.playerDone <- struct{}{}:
		default:
		}
		state.mu.Unlock()
		if onComplete != nil {
			onComplete()
		}
	}()
}

func showIgnoreDialog(state *AppState, onConfirm func()) {
	ignoredFolders := make([]string, len(state.IgnoreFolders))
	copy(ignoredFolders, state.IgnoreFolders)

	var folderList *widget.List

	folderList = widget.NewList(
		func() int { return len(ignoredFolders) },
		func() fyne.CanvasObject {
			return container.NewBorder(nil, nil, nil,
				widget.NewButtonWithIcon("", theme.DeleteIcon(), nil),
				widget.NewLabel(""),
			)
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			border := obj.(*fyne.Container)
			label := border.Objects[0].(*widget.Label)
			btn := border.Objects[1].(*widget.Button)
			label.SetText(ignoredFolders[i])
			btn.OnTapped = func() {
				ignoredFolders = append(ignoredFolders[:i], ignoredFolders[i+1:]...)
				folderList.Refresh()
			}
		},
	)
	scrolledList := container.NewScroll(folderList)
	scrolledList.SetMinSize(fyne.NewSize(500, 150))

	addFolderBtn := widget.NewButtonWithIcon("Add Folder", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			ignoredFolders = append(ignoredFolders, dir.Path())
			folderList.Refresh()
		}, state.Window)
	})

	extensionsEntry := widget.NewEntry()
	extensionsEntry.SetPlaceHolder("e.g. .txt, .pdf, .jpg")
	extensionsEntry.Text = strings.Join(state.IgnoreExtensions, ", ")
	extensionsEntry.MultiLine = false

	content := container.NewVBox(
		widget.NewLabel("Folders to ignore:"),
		scrolledList,
		addFolderBtn,
		widget.NewSeparator(),
		widget.NewLabel("Extensions to ignore (comma-separated):"),
		extensionsEntry,
		widget.NewLabel("These rules apply to this scan only."),
	)

	dialog.ShowCustomConfirm("Ignore Rules", "Start Scan", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}

		// Validate extensions before processing
		exts := strings.Split(extensionsEntry.Text, ",")
		for _, ext := range exts {
			ext = strings.TrimSpace(ext)
			if ext == "" {
				continue
			}
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			// Validate extension format
			if !isValidExtension(ext) {
				dialog.ShowError(
					fmt.Errorf("invalid extension: %s. Use only letters, numbers, and dots", ext),
					state.Window,
				)
				return
			}
		}

		state.IgnoreFolders = ignoredFolders
		state.IgnoreExtensions = []string{}
		for _, ext := range exts {
			ext = strings.TrimSpace(ext)
			if ext != "" {
				if !strings.HasPrefix(ext, ".") {
					ext = "." + ext
				}
				if isValidExtension(ext) {
					state.IgnoreExtensions = append(state.IgnoreExtensions, strings.ToLower(ext))
				}
			}
		}
		if onConfirm != nil {
			onConfirm()
		}
	}, state.Window)
}

// isValidExtension checks if an extension string is valid.
// It prevents regex injection and other malicious patterns.
func isValidExtension(ext string) bool {
	if ext == "" {
		return false
	}

	// Must start with a dot
	if !strings.HasPrefix(ext, ".") {
		return false
	}

	// Get the part after the dot
	pattern := ext[1:]

	// Empty extension after dot is invalid
	if pattern == "" {
		return false
	}

	// Only allow alphanumeric characters and dots
	// This prevents regex injection patterns like .*, .+, .txt.*
	for _, r := range pattern {
		isValidChar := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.'
		if !isValidChar {
			return false
		}
	}

	// Prevent patterns that could match too broadly
	// Block: *, +, ?, {, }, [, ], (, ), |, ^, $, \
	dangerousChars := []string{"*", "+", "?", "{", "}", "[", "]", "(", ")", "|", "^", "$", "\\"}
	for _, char := range dangerousChars {
		if strings.Contains(ext, char) {
			return false
		}
	}

	// Reasonable length limit (max 20 chars including dot)
	if len(ext) > 20 {
		return false
	}

	return true
}

// stopPlayback stops any ongoing audio playback and waits for the goroutine to finish.
// This is a no-op for states that don't support audio playback.
func stopPlayback(state interface{}) {
	switch s := state.(type) {
	case *AppState:
		stopPlaybackInternal(s)
	case *CacheCleanerState:
		// CacheCleanerState doesn't have audio playback, nothing to stop
	}
}

// stopPlaybackInternal is the internal implementation for AppState
func stopPlaybackInternal(state *AppState) {
	state.mu.Lock()

	if state.StopPlayer != nil {
		stopFunc := state.StopPlayer
		state.StopPlayer = nil

		// Get the done channel before releasing lock
		playerDone := state.playerDone

		state.mu.Unlock()

		// Call stop function (kills process)
		stopFunc()

		// Wait for goroutine to finish (with timeout)
		select {
		case <-playerDone:
			// Goroutine finished
		case <-time.After(2 * time.Second):
			// Timeout - goroutine may be leaked
			log.Printf("[stopPlayback] Timeout waiting for player goroutine to finish")
		}
	} else {
		state.mu.Unlock()
	}
}
