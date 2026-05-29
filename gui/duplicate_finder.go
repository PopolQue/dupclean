package gui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"dupclean/cleaner"
	"dupclean/gui/components"
	"dupclean/internal/fsutil"
	"dupclean/internal/trash"
	"dupclean/scanner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	moveToTrash    = trash.MoveToTrash
	findDuplicates = scanner.FindDuplicates
	goos           = runtime.GOOS
	userHomeDir    = os.UserHomeDir
	absPath        = filepath.Abs
	pathSeparator  = string(filepath.Separator)
	osRemoveAll    = os.RemoveAll
	safeToDelete   = func(f scanner.FileInfo) (bool, error) {
		info, err := os.Stat(f.Path)
		if err != nil {
			return false, err // file gone or inaccessible
		}
		if info.Size() != f.Size || !info.ModTime().Equal(f.ModTime) {
			return false, fmt.Errorf("file modified since scan")
		}
		return true, nil
	}
)

// DuplicateFinderWidget creates a duplicate finder UI component for a specific mode
func DuplicateFinderWidget(state *AppState, mode string) fyne.CanvasObject {
	state.mu.Lock()
	state.CurrentMode = mode
	state.mu.Unlock()

	title := "General Finder"
	subtitle := "Find exact byte-for-byte duplicates"
	if mode == "audio" {
		title = "Audio Finder"
		subtitle = "Find identical audio content regardless of file format"
	} else if mode == "photo" {
		title = "Photo Finder"
		subtitle = "Find visually similar or identical images"
	}

	// Options area (Cards)
	folderCard := createSelectionCard(state)

	// Mode-specific options
	var modeOptions *widget.Form
	if mode == "photo" {
		similaritySlider := widget.NewSlider(50, 100)
		similaritySlider.SetValue(90)
		similarityLabel := widget.NewLabel("90%")
		similaritySlider.OnChanged = func(v float64) {
			similarityLabel.SetText(fmt.Sprintf("%d%%", int(v)))
		}
		modeOptions = widget.NewForm(
			widget.NewFormItem("Similarity", container.NewBorder(nil, nil, nil, similarityLabel, similaritySlider)),
		)
	} else if mode == "audio" {
		modeOptions = widget.NewForm(
			widget.NewFormItem("Depth", widget.NewSelect([]string{"Fast", "Deep", "Exhaustive"}, func(s string) {})),
		)
	} else {
		modeOptions = widget.NewForm(
			widget.NewFormItem("Method", widget.NewSelect([]string{"MD5", "SHA256", "XXHash"}, func(s string) {})),
		)
	}

	// Scan settings
	scanHiddenCheck := widget.NewCheck("Scan hidden files", func(b bool) {})
	followSymlinksCheck := widget.NewCheck("Follow symlinks", func(b bool) {})
	scanSettings := container.NewHBox(scanHiddenCheck, followSymlinksCheck)

	optionsCard := widget.NewCard("Scan Settings", "Configure how we identify duplicates", container.NewVBox(
		modeOptions,
		scanSettings,
	))

	ignoreCard := createOptionsCard(state) // Existing ignore rules card

	options := container.NewVBox(folderCard, optionsCard, ignoreCard)

	// Action area (Start Button | Progress Bar)
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	state.progressComponents = &progressComponents{
		label:  widget.NewLabel(""),
		status: widget.NewLabel(""),
		bar:    progressBar,
	}

	scanBtn := createScanButton(state, folderCard, nil)
	registerStartButton(scanBtn)

	// Log/Action area
	logArea := container.NewVBox(
		container.NewHBox(state.progressComponents.label, layout.NewSpacer(), state.progressComponents.status),
	)

	return components.FixedTabLayout(
		title,
		subtitle,
		options,
		scanBtn,
		progressBar,
		logArea,
	)
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
		state.updateContent(DuplicateFinderWidget(state, state.CurrentMode))
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

	actionButtons := components.ActionFooter(
		cancelBtn,
		nil,
		container.NewHBox(smartBtn, cleanBtn),
	)

	return components.ToolPageWithFooter("Scan Results", statsText, groupDisplay, actionButtons)
}

// DuplicateNoResultsWidget creates the "no duplicates found" UI
func DuplicateNoResultsWidget(state *AppState) fyne.CanvasObject {
	statsText := fmt.Sprintf(
		"Scanned %d files in %s. Your files are clean!",
		state.Stats.TotalScanned,
		state.Stats.ScanDuration.Round(time.Second),
	)

	backBtn := widget.NewButtonWithIcon("Back to Home", theme.HomeIcon(), func() {
		state.updateContent(DuplicateFinderWidget(state, state.CurrentMode))
	})
	backBtn.Importance = widget.HighImportance

	return components.StatusPage(
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
		state.updateContent(DuplicateFinderWidget(state, state.CurrentMode))
	})
	backBtn.Importance = widget.HighImportance

	quitBtn := widget.NewButtonWithIcon("Quit", theme.CancelIcon(), func() {
		state.Window.Close()
	})

	btnRow := container.NewHBox(backBtn, quitBtn)

	return components.StatusPage(
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

func createSelectionCard(state *AppState) *widget.Card {
	picker := components.FolderPicker("Select a folder or paste path here...", state.FolderPath, true, state.Window, func(path string) {
		state.FolderPath = path
	})

	return widget.NewCard("Target Folder", "Select the directory you want to scan for duplicates", picker)
}

func createOptionsCard(state *AppState) *widget.Card {
	scanAllCheck := widget.NewCheck("Scan all file types (not just audio)", func(checked bool) {
		state.ScanAll = checked
	})

	ignoreBtn := widget.NewButtonWithIcon("Configure Ignore Rules", theme.SettingsIcon(), func() {
		showIgnoreDialog(state, nil)
	})
	ignoreBtn.Importance = widget.LowImportance

	content := container.NewVBox(
		scanAllCheck,
		ignoreBtn,
	)

	return widget.NewCard("Scan Options", "Configure filters and ignore rules", content)
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

	card := widget.NewCard("Status", "Current operation progress", content)

	// Store references in state for updates
	state.progressComponents = &progressComponents{
		label:  progressLabel,
		status: statusLabel,
		bar:    progressBar,
	}

	return card
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
	picker := folderCard.Content
	// Picker is a border container from components.FolderPicker
	// The entry is the first object (center)
	folderEntry := picker.(*fyne.Container).Objects[0].(*widget.Entry)
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

		groups, stats, err := findDuplicates(folderPath, scanAll, progressCallback, ignoreFolders, ignoreExtensions)
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

		// Sort groups by size (largest first)
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Files[0].Size > groups[j].Files[0].Size
		})

		// Initialize selections: keep first file by default after sorting
		state.mu.Lock()
		state.Groups = groups
		state.Stats = stats
		state.IsScanning = false

		state.Selections = make([][]bool, len(groups))
		for i := range groups {
			// Sort files within group: prefer shallowest path
			sort.Slice(groups[i].Files, func(a, b int) bool {
				da := strings.Count(groups[i].Files[a].Path, string(filepath.Separator))
				db := strings.Count(groups[i].Files[b].Path, string(filepath.Separator))
				if da != db {
					return da < db
				}
				return groups[i].Files[a].ModTime.Before(groups[i].Files[b].ModTime)
			})

			state.Selections[i] = make([]bool, len(groups[i].Files))
			if len(groups[i].Files) > 0 {
				state.Selections[i][0] = true
			}
		}
		state.mu.Unlock()

		ShowDuplicateResults(state)
	}()
}

func verifyDeletionSafety(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	abs, err := absPath(path)
	if err != nil {
		return err
	}

	// Don't allow deleting root or home directory
	if abs == "/" || abs == `\` {
		return fmt.Errorf("cannot delete root directory")
	}

	home, err := userHomeDir()
	if err == nil && abs == home {
		return fmt.Errorf("cannot delete home directory")
	}

	return nil
}

func cleanSelected(state *AppState) {
	stopPlayback(state)

	state.mu.Lock()
	groups := make([]scanner.DuplicateGroup, len(state.Groups))
	copy(groups, state.Groups)
	selections := make([][]bool, len(state.Selections))
	for i := range state.Selections {
		selections[i] = make([]bool, len(state.Selections[i]))
		copy(selections[i], state.Selections[i])
	}
	state.mu.Unlock()

	var deletedCount int
	var freedBytes int64
	var skippedCount int
	var skippedFiles []string

	for i, group := range groups {
		for j, f := range group.Files {
			if !selections[i][j] {
				// Safety check before deletion
				if err := verifyDeletionSafety(f.Path); err != nil {
					log.Printf("[GUI] Skipping protected path: %s (%v)", f.Path, err)
					skippedCount++
					skippedFiles = append(skippedFiles, f.Name)
					continue
				}

				if ok, err := safeToDelete(f); !ok {
					log.Printf("[cleanSelected] Skipping %s: %v", f.Path, err)
					skippedCount++
					skippedFiles = append(skippedFiles, f.Name)
					continue
				}
				if err := moveToTrash(f.Path); err == nil {
					deletedCount++
					freedBytes += f.Size
				}
			}
		}
	}

	state.mu.Lock()
	state.DeletedCount += deletedCount
	state.FreedBytes += freedBytes
	state.SkippedCount += skippedCount
	state.SkippedFiles = append(state.SkippedFiles, skippedFiles...)
	state.Groups = nil
	state.Selections = nil
	state.mu.Unlock()

	state.updateContent(DuplicateFinalWidget(state))
}

func createGroupDisplay(state *AppState) fyne.CanvasObject {
	accordion := widget.NewAccordion()

	state.mu.RLock()
	groups := state.Groups
	state.mu.RUnlock()

	for i, group := range groups {
		fileSize := fsutil.FormatBytes(group.Files[0].Size)
		title := fmt.Sprintf("%s (%d files, %s each)", group.Files[0].Name, len(group.Files), fileSize)

		// Create group content
		groupContent := container.NewVBox()

		for j, f := range group.Files {
			groupContent.Add(createFileCard(i, j, f, state))
		}

		item := widget.NewAccordionItem(title, groupContent)
		accordion.Append(item)
	}

	return accordion
}

func createFileCard(groupIndex, fileIndex int, f scanner.FileInfo, state *AppState) *widget.Card {
	// File number and name
	title := fmt.Sprintf("[%d] %s", fileIndex+1, f.Name)

	// Keep checkbox
	keepCheck := widget.NewCheck("Keep", func(checked bool) {
		state.mu.Lock()
		if groupIndex < len(state.Selections) && fileIndex < len(state.Selections[groupIndex]) {
			state.Selections[groupIndex][fileIndex] = checked
		}
		state.mu.Unlock()
	})
	state.mu.RLock()
	if groupIndex < len(state.Selections) && fileIndex < len(state.Selections[groupIndex]) {
		keepCheck.Checked = state.Selections[groupIndex][fileIndex]
	}
	state.mu.RUnlock()

	// Path with selectable entry for better display
	pathEntry := widget.NewEntry()
	pathEntry.SetText(f.Path)
	pathEntry.Disable()
	pathEntry.Wrapping = fyne.TextWrapBreak
	pathEntry.MultiLine = true

	// Metadata
	metadata := fmt.Sprintf("Size: %s  •  Modified: %s", fsutil.FormatBytes(f.Size), f.ModTime.Format("2006-01-02 15:04"))

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

	return components.ResultCard(title, pathEntry, metadata, keepCheck, playBtn)
}

func keepAndDelete(state *AppState, keepIndex int, files []scanner.FileInfo) {
	stopPlayback(state)

	state.mu.Lock()
	defer state.mu.Unlock()

	keepAndDeleteLocked(state, keepIndex, files)
}

// keepAndDeleteLocked performs the deletion and state update.
// The caller MUST hold state.mu.Lock().
func keepAndDeleteLocked(state *AppState, keepIndex int, files []scanner.FileInfo) {
	if len(files) == 0 {
		return
	}
	groupHash := files[0].Hash

	for idx, f := range files {
		if idx == keepIndex {
			continue
		}

		if ok, err := safeToDelete(f); !ok {
			log.Printf("[keepAndDeleteLocked] Skipping %s: %v", f.Path, err)
			state.SkippedCount++
			state.SkippedFiles = append(state.SkippedFiles, f.Name)
			continue
		}

		if err := moveToTrash(f.Path); err == nil {
			state.DeletedCount++
			state.FreedBytes += f.Size
		} else {
			log.Printf("[keepAndDeleteLocked] Failed to trash %s: %v", f.Path, err)
			state.SkippedCount++
			state.SkippedFiles = append(state.SkippedFiles, f.Name)
		}
	}

	// Remove the resolved group from the list and sync selections

	for i, g := range state.Groups {
		if g.Hash == groupHash {
			state.Groups = append(state.Groups[:i], state.Groups[i+1:]...)
			if i < len(state.Selections) {
				state.Selections = append(state.Selections[:i], state.Selections[i+1:]...)
			}
			break
		}
	}
}

// SmartCleanAll automatically resolves all duplicate groups by keeping the "best" file
func SmartCleanAll(state *AppState) {
	stopPlayback(state)

	state.mu.Lock()
	// Copy groups to avoid iteration issues while modifying state.Groups
	groups := make([]scanner.DuplicateGroup, len(state.Groups))
	copy(groups, state.Groups)

	for _, g := range groups {
		files := g.Files
		if len(files) < 2 {
			continue
		}

		// Selection logic: prefer shallowest path, then oldest modification time
		sort.Slice(files, func(i, j int) bool {
			di := strings.Count(files[i].Path, pathSeparator)
			dj := strings.Count(files[j].Path, pathSeparator)
			if di != dj {
				return di < dj
			}
			return files[i].ModTime.Before(files[j].ModTime)
		})

		// Keep index 0 (the best match based on sort above)
		keepAndDeleteLocked(state, 0, files)
	}
	state.mu.Unlock()

	if state.RefreshResults != nil {
		state.RefreshResults()
	}
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

	// Validation label for extension input
	extValidationLabel := widget.NewLabel("")
	extValidationLabel.Hide()

	content := container.NewVBox(
		widget.NewLabel("Folders to ignore:"),
		scrolledList,
		addFolderBtn,
		widget.NewSeparator(),
		widget.NewLabel("Extensions to ignore (comma-separated):"),
		extensionsEntry,
		extValidationLabel,
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
				// Validate extension - reject wildcards and dangerous patterns
				if isValidExtension(ext) {
					state.IgnoreExtensions = append(state.IgnoreExtensions, strings.ToLower(ext))
				} else {
					extValidationLabel.SetText("Invalid extension ignored: " + ext + " (wildcards not allowed)")
					extValidationLabel.Show()
				}
			}
		}
		if onConfirm != nil {
			onConfirm()
		}
	}, state.Window)
}

// isValidExtension validates file extensions to prevent dangerous patterns
func isValidExtension(ext string) bool {
	if ext == "" {
		return false
	}
	if !strings.HasPrefix(ext, ".") {
		return false
	}
	if len(ext) == 1 { // only dot
		return false
	}
	if len(ext) > 20 {
		return false
	}
	// Reject wildcards and dangerous patterns
	return !strings.ContainsAny(ext, "*?+{}[]()|^$\\")
}

func stopPlayback(state *AppState) {
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
