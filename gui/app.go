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
	Window            fyne.Window
	FolderPath        string
	ScanAll           bool
	IsScanning        bool
	ProgressText      string
	ProgressValue     float64
	Groups            []scanner.DuplicateGroup
	CurrentGroupIndex int
	DeletedCount      int
	FreedBytes        int64
	Stats             scanner.ScanStats
	CurrentPlayer     *exec.Cmd
	StopPlayer        func()
	IgnoreFolders     []string
	IgnoreExtensions  []string
}

func RunGUI() {
	log.Println("RunGUI: starting...")

	fyneApp := app.New()

	log.Println("RunGUI: setting icon...")
	fyneApp.SetIcon(theme.FolderOpenIcon())

	log.Println("RunGUI: creating window...")
	w := fyneApp.NewWindow("DupClean - Audio Duplicate Finder")
	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(true)

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
	title := canvas.NewText("DupClean - Audio Duplicate Finder", theme.PrimaryColor())
	title.TextSize = 24
	title.Alignment = fyne.TextAlignCenter

	folderLabel := widget.NewLabel("Select folder to scan:")
	folderEntry := widget.NewEntry()
	folderEntry.Disable()
	browseBtn := widget.NewButton("Browse...", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			folderEntry.SetText(dir.Path())
			state.FolderPath = dir.Path()
		}, state.Window)
	})

	folderRow := container.NewBorder(nil, nil, folderLabel, browseBtn, folderEntry)

	scanAllCheck := widget.NewCheck("Scan all file types (not just audio)", func(checked bool) {
		state.ScanAll = checked
	})

	progressBar := widget.NewProgressBar()
	progressLabel := widget.NewLabel("Ready")

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
			startScan(state, folderEntry, progressBar, progressLabel)
		})
	})
	scanBtn.Disable()

	folderEntry.OnChanged = func(text string) {
		state.FolderPath = text
		if text != "" {
			scanBtn.Enable()
		} else {
			scanBtn.Disable()
		}
	}

	progressContainer := container.NewVBox(progressBar, progressLabel)

	btnContainer := container.NewVBox(
		folderRow,
		scanAllCheck,
		scanBtn,
		layout.NewSpacer(),
		progressContainer,
	)

	return container.NewVBox(
		title,
		layout.NewSpacer(),
		btnContainer,
	)
}

func startScan(state *AppState, _ *widget.Entry, progressBar *widget.ProgressBar, progressLabel *widget.Label) {
	state.IsScanning = true
	progressLabel.SetText("Scanning folder...")
	progressBar.SetValue(0)

	go func() {
		progressCallback := func(progress scanner.ScanProgress) {
			fyne.Do(func() {
				progressBar.SetValue(progress.Percent)
				progressLabel.SetText(fmt.Sprintf("[%d%%] %s", int(progress.Percent*100), progress.Phase))
			})
		}
		groups, stats, err := scanner.FindDuplicates(state.FolderPath, state.ScanAll, progressCallback, state.IgnoreFolders, state.IgnoreExtensions)
		if err != nil {
			state.IsScanning = false
			fyne.Do(func() {
				progressLabel.SetText(fmt.Sprintf("Error: %v", err))
			})
			return
		}

		fyne.Do(func() {
			progressLabel.SetText(fmt.Sprintf("Found %d duplicate groups", len(groups)))
			progressBar.SetValue(1)
		})

		state.Groups = groups
		state.IsScanning = false

		state.Window.Content().Refresh()
		showResults(state, stats)
	}()
}

func showResults(state *AppState, stats scanner.ScanStats) {
	state.Stats = stats
	if len(state.Groups) == 0 {
		state.Window.SetContent(createNoDuplicatesUI(state, state.Stats))
		return
	}
	state.Window.SetContent(createResultsUI(state, state.Stats))
}

func createNoDuplicatesUI(state *AppState, stats scanner.ScanStats) fyne.CanvasObject {
	title := canvas.NewText("No Duplicates Found!", theme.PrimaryColor())
	title.TextSize = 24
	title.Alignment = fyne.TextAlignCenter

	statsLabel := widget.NewLabel(fmt.Sprintf(
		"Scanned %d files in %s\nNo duplicate audio files were found.",
		stats.TotalScanned,
		stats.ScanDuration.Round(100000*time.Microsecond),
	))
	statsLabel.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButton("Back to Home", func() {
		state.Window.SetContent(createMainUI(state))
	})

	return container.NewVBox(
		title,
		layout.NewSpacer(),
		statsLabel,
		layout.NewSpacer(),
		backBtn,
		layout.NewSpacer(),
	)
}

func createResultsUI(state *AppState, stats scanner.ScanStats) fyne.CanvasObject {
	title := canvas.NewText("Scan Results", theme.PrimaryColor())
	title.TextSize = 24
	title.Alignment = fyne.TextAlignCenter

	statsText := fmt.Sprintf(
		"Found %d duplicate groups | %d extra copies | %s wasted",
		len(state.Groups),
		stats.TotalDupes,
		formatBytes(stats.WastedBytes),
	)
	statsLabel := widget.NewLabel(statsText)
	statsLabel.Alignment = fyne.TextAlignCenter

	groupsContainer := container.NewVBox()

	state.CurrentGroupIndex = 0

	scroll := container.NewScroll(groupsContainer)
	scroll.SetMinSize(fyne.NewSize(700, 400))

	updateGroupDisplay := func() {
		groupsContainer.Objects = nil

		if state.CurrentGroupIndex >= len(state.Groups) {
			if len(state.Groups) == 0 {
				state.Window.SetContent(createFinalUI(state))
				return
			}
			state.CurrentGroupIndex = len(state.Groups) - 1
		}

		group := state.Groups[state.CurrentGroupIndex]
		groupNum := state.CurrentGroupIndex + 1
		totalGroups := len(state.Groups)

		groupLabel := canvas.NewText(
			fmt.Sprintf("Group %d of %d (%s)", groupNum, totalGroups, formatBytes(group.Files[0].Size)),
			theme.PrimaryColor(),
		)
		groupLabel.TextSize = 18
		groupsContainer.AddObject(groupLabel)
		groupsContainer.AddObject(widget.NewSeparator())

		files := group.Files
		sort.Slice(files, func(i, j int) bool {
			di := strings.Count(files[i].Path, string(filepath.Separator))
			dj := strings.Count(files[j].Path, string(filepath.Separator))
			if di != dj {
				return di < dj
			}
			return files[i].ModTime.Before(files[j].ModTime)
		})

		for idx, f := range files {
			row := createFileRow(idx+1, f, state, stats)
			groupsContainer.AddObject(row)
		}

		btnRow := container.NewHBox(
			widget.NewButton("Skip Group", func() {
				state.CurrentGroupIndex++
				state.Window.SetContent(createResultsUI(state, state.Stats))
			}),
			widget.NewButton("Skip All", func() {
				state.Window.SetContent(createFinalUI(state))
			}),
		)

		if len(files) > 0 {
			keepBtn := widget.NewButton("Keep #1 & Delete Others", func() {
				keepAndDelete(state, 0, files)
				if len(state.Groups) == 0 {
					state.Window.SetContent(createFinalUI(state))
				} else {
					state.Window.SetContent(createResultsUI(state, state.Stats))
				}
			})
			btnRow.AddObject(keepBtn)
		}

		groupsContainer.AddObject(btnRow)
		groupsContainer.AddObject(widget.NewSeparator())

		scroll.Refresh()
	}

	updateGroupDisplay()

	return container.NewVBox(
		title,
		statsLabel,
		layout.NewSpacer(),
		scroll,
	)
}

func createFileRow(num int, f scanner.FileInfo, state *AppState, _ scanner.ScanStats) fyne.CanvasObject {
	card := widget.NewCard(fmt.Sprintf("[%d] %s", num, f.Name), f.Path,
		container.NewVBox(
			widget.NewLabel(fmt.Sprintf("Size: %s | Modified: %s", formatBytes(f.Size), f.ModTime.Format("2006-01-02 15:04"))),
		))

	playBtn := widget.NewButtonWithIcon("", theme.MediaPlayIcon(), func() {
		playFile(state, f.Path)
	})

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Delete File?", fmt.Sprintf("Move '%s' to Trash?", f.Name), func(ok bool) {
			if ok {

				if state.CurrentPlayer != nil {
					stopPlayback(state)
					state.CurrentPlayer = nil
				}

				moveToTrash(f.Path)
				state.DeletedCount++
				state.FreedBytes += f.Size

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

				state.Window.SetContent(createResultsUI(state, state.Stats))
			}
		}, state.Window)
	})
	deleteBtn.Importance = widget.DangerImportance
	buttons := container.NewHBox(playBtn, deleteBtn)
	return container.NewBorder(nil, nil, card, buttons)
}

func keepAndDelete(state *AppState, keepIndex int, files []scanner.FileInfo) {
	stopPlayback(state)
	for idx, f := range files {
		if idx == keepIndex {
			continue
		}
		moveToTrash(f.Path)
		state.DeletedCount++
		state.FreedBytes += f.Size
	}

	i := state.CurrentGroupIndex
	state.Groups = append(state.Groups[:i], state.Groups[i+1:]...)
}

func createFinalUI(state *AppState) fyne.CanvasObject {
	title := canvas.NewText("Complete!", theme.PrimaryColor())
	title.TextSize = 24
	title.Alignment = fyne.TextAlignCenter

	var message string
	if state.DeletedCount == 0 {
		message = "No files were deleted. Your files are safe."
	} else {
		message = fmt.Sprintf("Moved %d file(s) to Trash\nFreed %s", state.DeletedCount, formatBytes(state.FreedBytes))
	}

	resultLabel := widget.NewLabel(message)
	resultLabel.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButton("Start New Scan", func() {
		state.Groups = nil
		state.CurrentGroupIndex = 0
		state.DeletedCount = 0
		state.FreedBytes = 0
		state.FolderPath = ""
		state.Window.SetContent(createMainUI(state))
	})

	quitBtn := widget.NewButton("Quit", func() {
		state.Window.Close()
	})

	btnRow := container.NewHBox(backBtn, quitBtn)

	return container.NewVBox(
		title,
		layout.NewSpacer(),
		resultLabel,
		layout.NewSpacer(),
		btnRow,
		layout.NewSpacer(),
	)
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

func playFile(state *AppState, path string) {
	if state.StopPlayer != nil {
		state.StopPlayer()
		state.StopPlayer = nil
		state.CurrentPlayer = nil
	}

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
	state.StopPlayer = func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	go func() {
		cmd.Run()
		if state.CurrentPlayer == cmd {
			state.CurrentPlayer = nil
			state.StopPlayer = nil
		}
	}()
}

func showIgnoreDialog(state *AppState, onConfirm func()) {
	ignoredFolders := []string{}

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

	content := container.NewVBox(
		widget.NewLabel("Folders to ignore:"),
		scrolledList,
		addFolderBtn,
		widget.NewSeparator(),
		widget.NewLabel("Extensions to ignore (comma-separated):"),
		extensionsEntry,
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
		onConfirm()
	}, state.Window)
}

func stopPlayback(state *AppState) {
	if state.StopPlayer != nil {
		state.StopPlayer()
		state.StopPlayer = nil
		state.CurrentPlayer = nil
	}
}
