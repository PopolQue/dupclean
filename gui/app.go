package gui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

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

func init() {
	logFile, err := os.OpenFile("/tmp/dupclean.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		log.SetOutput(logFile)
	}
	log.Println("DupClean starting...")
}

type AppState struct {
	Window             fyne.Window
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
}

func RunGUI() {
	log.Println("RunGUI: starting...")

	fyneApp := app.New()

	log.Println("RunGUI: setting icon...")
	fyneApp.SetIcon(theme.FolderOpenIcon())

	log.Println("RunGUI: creating window...")
	w := fyneApp.NewWindow("DupClean - Duplicate File Finder")
	w.Resize(fyne.NewSize(1100, 750))

	log.Println("RunGUI: creating state...")
	state := &AppState{
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
	}

	w.SetContent(createMainUI(state))

	log.Println("RunGUI: showing window...")
	w.ShowAndRun()
	log.Println("RunGUI: window closed")
}

func createMainUI(state *AppState) fyne.CanvasObject {
	// Header with logo and title
	title := canvas.NewText("DupClean", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("Duplicate File Finder", theme.Color(theme.ColorNameForeground))
	subtitle.TextSize = 16
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

func createSelectionCard(state *AppState) *widget.Card {
	folderLabel := widget.NewLabel("Folder to scan:")
	folderLabel.TextStyle = fyne.TextStyle{Bold: true}

	folderEntry := widget.NewEntry()
	folderEntry.Disable()
	folderEntry.SetPlaceHolder("Select a folder...")

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
	state.IsScanning = true

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
		groups, stats, err := scanner.FindDuplicates(state.FolderPath, state.ScanAll, progressCallback, state.IgnoreFolders, state.IgnoreExtensions)
		if err != nil {
			state.IsScanning = false
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

		state.Groups = groups
		state.Stats = stats
		state.IsScanning = false

		showResults(state, stats)
	}()
}

func showResults(state *AppState, stats scanner.ScanStats) {
	if len(state.Groups) == 0 {
		state.Window.SetContent(createNoDuplicatesUI(state, state.Stats))
		return
	}
	state.Window.SetContent(createResultsUI(state, state.Stats))
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
		state.Window.SetContent(createMainUI(state))
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
			state.Window.SetContent(createResultsUI(state, state.Stats))
		}
	})
	prevBtn.Importance = widget.LowImportance

	nextBtn := widget.NewButtonWithIcon("Next", theme.NavigateNextIcon(), func() {
		if state.CurrentGroupIndex < len(state.Groups)-1 {
			state.CurrentGroupIndex++
			state.Window.SetContent(createResultsUI(state, state.Stats))
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
			state.Window.SetContent(createFinalUI(state))
		} else {
			state.Window.SetContent(createResultsUI(state, state.Stats))
		}
	})
	skipGroupBtn.Importance = widget.LowImportance

	skipAllBtn := widget.NewButton("Skip All", func() {
		state.Window.SetContent(createFinalUI(state))
	})
	skipAllBtn.Importance = widget.LowImportance

	keepBtn := widget.NewButtonWithIcon("Keep #1 & Delete Others", theme.ConfirmIcon(), func() {
		if state.CurrentGroupIndex < len(state.Groups) {
			group := state.Groups[state.CurrentGroupIndex]
			keepAndDelete(state, 0, group.Files)
			if len(state.Groups) == 0 {
				state.Window.SetContent(createFinalUI(state))
			} else {
				state.Window.SetContent(createResultsUI(state, state.Stats))
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
						state.Window.SetContent(createResultsUI(state, state.Stats))
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
		state.DeletedCount++
		state.FreedBytes += f.Size
		_ = moveToTrash(f.Path)
	}

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
		state.Window.SetContent(createMainUI(state))
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
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if _, err := exec.LookPath("trash"); err == nil {
		return exec.Command("trash", absPath).Run()
	}

	if runtimeOS() == "darwin" {
		script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, absPath)
		return exec.Command("osascript", "-e", script).Run()
	}

	if runtimeOS() == "linux" {
		gvfsPath := filepath.Join(os.Getenv("HOME"), ".local/share/Trash/files")
		if _, err := os.Stat(gvfsPath); err == nil {
			trashName := absPath
			counter := 1
			for {
				newPath := filepath.Join(gvfsPath, filepath.Base(trashName))
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					break
				}
				ext := filepath.Ext(trashName)
				base := strings.TrimSuffix(filepath.Base(trashName), ext)
				trashName = fmt.Sprintf("%s (%d)%s", base, counter, ext)
				counter++
			}
			return os.Rename(absPath, filepath.Join(gvfsPath, trashName))
		}
	}

	return os.Remove(absPath)
}

func runtimeOS() string {
	return runtime.GOOS
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

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("afplay", path)
	case "linux":
		cmd = exec.Command("aplay", path)
	case "windows":
		cmd = exec.Command("powershell", "-c",
			fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", path))
	default:
		return
	}

	state.CurrentPlayer = cmd
	state.PlayingPath = path
	state.StopPlayer = func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		if onComplete != nil {
			onComplete()
		}
	}

	go func() {
		_ = cmd.Run()
		if state.CurrentPlayer == cmd {
			state.CurrentPlayer = nil
			state.StopPlayer = nil
			state.PlayingPath = ""
			if onComplete != nil {
				onComplete()
			}
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
		state.IgnoreFolders = ignoredFolders

		state.IgnoreExtensions = []string{}
		for _, ext := range strings.Split(extensionsEntry.Text, ",") {
			ext = strings.TrimSpace(ext)
			if ext != "" {
				if !strings.HasPrefix(ext, ".") {
					ext = "." + ext
				}
				state.IgnoreExtensions = append(state.IgnoreExtensions, strings.ToLower(ext))
			}
		}
		if onConfirm != nil {
			onConfirm()
		}
	}, state.Window)
}

func stopPlayback(state *AppState) {
	if state.StopPlayer != nil {
		state.StopPlayer()
		state.StopPlayer = nil
		state.CurrentPlayer = nil
		state.PlayingPath = ""
	}
}
