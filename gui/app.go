package gui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"dupclean/scanner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type AppState struct {
	Window            fyne.Window
	FolderPath        binding.String
	ScanAll           binding.Bool
	IsScanning        binding.Bool
	ProgressText      binding.String
	ProgressValue     binding.Float
	Groups            []scanner.DuplicateGroup
	CurrentGroupIndex int
	DeletedCount      int
	FreedBytes        int64
}

func RunGUI() {
	app := fyne.CurrentApp()
	app.SetIcon(theme.FolderOpenIcon())

	w := app.NewWindow("DupClean - Audio Duplicate Finder")
	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(true)

	state := &AppState{
		Window:        w,
		FolderPath:    binding.NewString(),
		ScanAll:       binding.NewBool(),
		IsScanning:    binding.NewBool(),
		ProgressText:  binding.NewString(),
		ProgressValue: binding.NewFloat(),
	}

	state.FolderPath.Set("")
	state.ProgressText.Set("Ready")

	w.SetContent(createMainUI(state))

	w.Show()
}

func createMainUI(state *AppState) fyne.CanvasObject {
	title := canvas.NewText("DupClean - Audio Duplicate Finder", theme.PrimaryColor())
	title.TextSize = 24
	title.Alignment = fyne.TextAlignCenter

	folderLabel := widget.NewLabel("Select folder to scan:")
	folderEntry := widget.NewEntryWithData(state.FolderPath)
	folderEntry.Disable()
	browseBtn := widget.NewButton("Browse...", func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableItem, err error) {
			if err != nil || dir == nil {
				return
			}
			state.FolderPath.Set(dir.Name())
		}, state.Window)
	})

	folderRow := container.NewBorder(nil, nil, folderLabel, browseBtn, folderEntry)

	scanAllCheck := widget.NewCheckWithData("Scan all file types (not just audio)", state.ScanAll)

	scanBtn := widget.NewButtonWithIcon("Start Scan", theme.SearchIcon(), func() {
		startScan(state)
	})
	scanBtn.Disable()

	state.FolderPath.AddListener(binding.NewDataListener(func() {
		folder, _ := state.FolderPath.Get()
		scanBtn.Disable()
		if folder != "" {
			scanBtn.Enable()
		}
	}))

	progressBar := widget.NewProgressBarWithData(state.ProgressValue)
	progressLabel := widget.NewLabelWithData(state.ProgressText)

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

func startScan(state *AppState) {
	folder, _ := state.FolderPath.Get()
	if folder == "" {
		dialog.ShowError(fmt.Errorf("please select a folder"), state.Window)
		return
	}

	info, err := os.Stat(folder)
	if err != nil || !info.IsDir() {
		dialog.ShowError(fmt.Errorf("invalid folder"), state.Window)
		return
	}

	scanAll, _ := state.ScanAll.Get()
	state.IsScanning.Set(true)
	state.ProgressText.Set("Scanning folder...")
	state.ProgressValue.Set(0)

	go func() {
		groups, stats, err := scanner.FindDuplicates(folder, scanAll)
		if err != nil {
			state.IsScanning.Set(false)
			state.ProgressText.Set(fmt.Sprintf("Error: %v", err))
			return
		}

		state.ProgressText.Set(fmt.Sprintf("Found %d duplicate groups", len(groups)))
		state.ProgressValue.Set(1)

		state.Groups = groups
		state.IsScanning.Set(false)

		showResults(state, stats)
	}()
}

func showResults(state *AppState, stats scanner.ScanStats) {
	state.Window.Content().Refresh()

	if len(state.Groups) == 0 {
		state.Window.SetContent(createNoDuplicatesUI(state, stats))
		return
	}

	state.Window.SetContent(createResultsUI(state, stats))
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
			state.Window.SetContent(createFinalUI(state))
			return
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
			row := createFileRow(idx+1, f, state)
			groupsContainer.AddObject(row)
		}

		btnRow := container.NewHBox(
			widget.NewButton("Skip Group", func() {
				state.CurrentGroupIndex++
				updateGroupDisplay()
			}),
			widget.NewButton("Skip All", func() {
				state.Window.SetContent(createFinalUI(state))
			}),
		)

		if len(files) > 0 {
			keepBtn := widget.NewButton(fmt.Sprintf("Keep #1 & Delete Others", len(files)), func() {
				keepAndDelete(state, 0, files)
				state.CurrentGroupIndex++
				updateGroupDisplay()
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

func createFileRow(num int, f scanner.FileInfo, state *AppState) fyne.CanvasObject {
	card := widget.NewCard()
	card.SetContent(container.NewVBox(
		widget.NewLabel(fmt.Sprintf("[%d] %s", num, f.Name)),
		widget.NewLabel(f.Path),
		widget.NewLabel(fmt.Sprintf("Size: %s | Modified: %s", formatBytes(f.Size), f.ModTime.Format("2006-01-02 15:04"))),
	))

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		dialog.ShowConfirm("Delete File?", fmt.Sprintf("Move '%s' to Trash?", f.Name), func(ok bool) {
			if ok {
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

				state.Window.SetContent(createResultsUI(state, scanner.ScanStats{}))
			}
		}, state.Window)
	})
	deleteBtn.Importance = widget.DangerImportance

	return container.NewBorder(nil, nil, card, deleteBtn)
}

func keepAndDelete(state *AppState, keepIndex int, files []scanner.FileInfo) {
	_ = files[keepIndex]
	for idx, f := range files {
		if idx == keepIndex {
			continue
		}
		moveToTrash(f.Path)
		state.DeletedCount++
		state.FreedBytes += f.Size
	}
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
		state.FolderPath.Set("")
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
	return strings.ToLower(os.Getenv("GOOS"))
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

var _ = strconv.Itoa
