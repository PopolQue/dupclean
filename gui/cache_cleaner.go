package gui

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"dupclean/cleaner"
	"dupclean/internal/fsutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CacheCleanerState holds the state for the cache cleaner widget
type CacheCleanerState struct {
	Window                 fyne.Window
	ContentContainer       *fyne.Container // Reference to content area
	Targets                []*cleaner.CleanTarget
	SelectedTargets        map[string]bool
	TotalSize              int64
	IsScanning             bool
	IsCleaning             bool
	CleanedCount           int
	CleanedBytes           int64
	MinAgeStr              string
	Concurrency            int
	cacheCleanerComponents *cacheCleanerComponents
}

type cacheCleanerComponents struct {
	scanBtn       *widget.Button
	progressLabel *widget.Label
	progressBar   *widget.ProgressBar
	minAgeEntry   *widget.Entry
	workersSelect *widget.Select
}

// updateContent updates the content container (preserves sidebar)
func (state *CacheCleanerState) updateContent(content fyne.CanvasObject) {
	if state.ContentContainer != nil {
		state.ContentContainer.Objects = []fyne.CanvasObject{content}
		state.ContentContainer.Refresh()
	}
}

// CacheCleanerWidget creates the cache cleaner UI component
func CacheCleanerWidget(state *CacheCleanerState) fyne.CanvasObject {
	// Options
	minAgeEntry := widget.NewEntry()
	minAgeEntry.SetPlaceHolder("e.g. 24h, 7d")
	minAgeEntry.SetText(state.MinAgeStr)
	minAgeEntry.OnChanged = func(s string) {
		state.MinAgeStr = s
	}

	workersSelect := widget.NewSelect([]string{"1", "2", "4", "8", "16"}, func(s string) {
		val, _ := strconv.Atoi(s)
		state.Concurrency = val
	})
	workersSelect.SetSelected(fmt.Sprintf("%d", state.Concurrency))

	optionsForm := widget.NewForm(
		widget.NewFormItem("Min Age", minAgeEntry),
		widget.NewFormItem("Workers", workersSelect),
	)

	// Scan button
	scanBtn := widget.NewButtonWithIcon("Scan for Caches", theme.SearchIcon(), func() {
		startCacheScan(state)
	})
	scanBtn.Importance = widget.HighImportance

	// Progress indicator
	progressLabel := widget.NewLabel("Ready to scan")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	progressCard := widget.NewCard("", "", container.NewVBox(
		progressLabel,
		progressBar,
	))

	// Disclaimer
	disclaimer := canvas.NewText("⚠️ Actual freed space may vary - cache files change constantly", theme.Color(theme.ColorNameWarning))
	disclaimer.TextSize = 12
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}
	disclaimer.Alignment = fyne.TextAlignCenter

	body := container.NewVBox(
		widget.NewCard("Options", "Filter and performance settings", optionsForm),
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), scanBtn, layout.NewSpacer()),
		progressCard,
		layout.NewSpacer(),
		disclaimer,
	)

	// Store references for updates
	state.cacheCleanerComponents = &cacheCleanerComponents{
		scanBtn:       scanBtn,
		progressLabel: progressLabel,
		progressBar:   progressBar,
		minAgeEntry:   minAgeEntry,
		workersSelect: workersSelect,
	}

	return createToolPage("Cache Cleaner", "Clean system, browser, and application caches", body)
}

func startCacheScan(state *CacheCleanerState) {
	state.IsScanning = true

	comp := state.cacheCleanerComponents
	comp.scanBtn.Disable()
	comp.minAgeEntry.Disable()
	comp.workersSelect.Disable()

	comp.progressLabel.SetText("Scanning for caches...")
	comp.progressBar.Show()
	comp.progressBar.SetValue(0)

	go func() {
		// Parse min age
		minAge, err := parseDuration(state.MinAgeStr)
		if err != nil {
			fyne.Do(func() {
				comp.progressLabel.SetText(fmt.Sprintf("Invalid Min Age: %v", err))
				comp.scanBtn.Enable()
				comp.minAgeEntry.Enable()
				comp.workersSelect.Enable()
			})
			return
		}

		// Get all cache targets
		targets := cleaner.Registry()

		// Filter out protected system directories
		filteredTargets := make([]*cleaner.CleanTarget, 0)
		for _, t := range targets {
			// Filter paths for this target
			cleanPaths := make([]string, 0)
			for _, path := range t.Paths {
				if !isProtectedPath(path) {
					cleanPaths = append(cleanPaths, path)
				}
			}

			// Only include target if it has cleanable paths
			if len(cleanPaths) > 0 {
				t.Paths = cleanPaths
				filteredTargets = append(filteredTargets, t)
			}
		}
		targets = filteredTargets

		// Scan each target
		opts := cleaner.ScanOptions{
			Concurrency: state.Concurrency,
			MinAge:      minAge,
			OnProgress: func(progress cleaner.Progress) {
				fyne.Do(func() {
					if progress.Total > 0 {
						comp.progressBar.SetValue(float64(progress.Done) / float64(progress.Total))
					}
					comp.progressLabel.SetText(fmt.Sprintf("Scanning: %s", progress.Current))
				})
			},
		}

		result, err := cleaner.Scan(targets, opts)
		if err != nil {
			fyne.Do(func() {
				comp.progressLabel.SetText(fmt.Sprintf("Error: %v", err))
				comp.scanBtn.Enable()
				comp.minAgeEntry.Enable()
				comp.workersSelect.Enable()
			})
			return
		}

		state.Targets = result.Targets
		state.TotalSize = result.TotalSize
		state.IsScanning = false

		// Pre-select safe targets
		state.SelectedTargets = make(map[string]bool)
		for _, t := range result.Targets {
			if t.TotalSize > 0 && (t.Risk == cleaner.RiskSafe || t.Risk == cleaner.RiskLow) {
				state.SelectedTargets[t.ID] = true
			}
		}

		fyne.Do(func() {
			displayCacheResults(state)
		})
	}()
}

func displayCacheResults(state *CacheCleanerState) {
	resultsContainer := container.NewVBox()

	// Group by category
	categories := make(map[string][]*cleaner.CleanTarget)
	for _, t := range state.Targets {
		if t.TotalSize > 0 {
			categories[t.Category] = append(categories[t.Category], t)
		}
	}

	// Sort categories
	catNames := make([]string, 0, len(categories))
	for cat := range categories {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	// Action buttons
	cleanBtn := widget.NewButtonWithIcon("Clean Selected", theme.ConfirmIcon(), func() {
		startCacheClean(state)
	})
	cleanBtn.Importance = widget.HighImportance

	// Create total label - will be updated on selection
	totalLabel := canvas.NewText("Selected: 0 B", theme.Color(theme.ColorNamePrimary))
	totalLabel.TextSize = 18
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Create category sections
	for _, cat := range catNames {
		targets := categories[cat]

		catLabel := canvas.NewText(cat, theme.Color(theme.ColorNamePrimary))
		catLabel.TextSize = 18
		catLabel.TextStyle = fyne.TextStyle{Bold: true}
		resultsContainer.Add(catLabel)
		resultsContainer.Add(widget.NewSeparator())

		for _, t := range targets {
			targetCard := createCacheTargetCard(state, t, totalLabel, cleanBtn)
			resultsContainer.Add(targetCard)
		}

		resultsContainer.Add(layout.NewSpacer())
	}

	updateCacheTotal(state, totalLabel, cleanBtn)

	cancelBtn := widget.NewButton("Cancel", func() {
		state.updateContent(CacheCleanerWidget(state))
	})
	cancelBtn.Importance = widget.LowImportance

	buttonRow := container.NewHBox(cancelBtn, layout.NewSpacer(), totalLabel, layout.NewSpacer(), cleanBtn)

	body := container.NewVBox(
		resultsContainer,
		widget.NewSeparator(),
		buttonRow,
	)

	subtitle := fmt.Sprintf("Found %s of cleanable caches", fsutil.FormatBytes(state.TotalSize))
	state.updateContent(createToolPage("Scan Results", subtitle, body))
}

// createCacheTargetCard creates a card for a cache target
func createCacheTargetCard(state *CacheCleanerState, target *cleaner.CleanTarget, totalLabel *canvas.Text, cleanBtn *widget.Button) *widget.Card {
	// Risk icon
	riskIcon := theme.ConfirmIcon()
	switch target.Risk {
	case cleaner.RiskModerate:
		riskIcon = theme.WarningIcon()
	case cleaner.RiskHigh:
		riskIcon = theme.ErrorIcon()
	}

	// Target checkbox
	targetCheck := widget.NewCheck(fmt.Sprintf("%s (%s)", target.Label, fsutil.FormatBytes(target.TotalSize)), func(checked bool) {
		state.SelectedTargets[target.ID] = checked
		updateCacheTotal(state, totalLabel, cleanBtn)
	})
	targetCheck.SetChecked(state.SelectedTargets[target.ID])

	// Description
	descLabel := widget.NewLabel(target.Description)
	descLabel.TextStyle = fyne.TextStyle{Italic: true}

	// File count
	fileCountLabel := widget.NewLabel(fmt.Sprintf("%d items", target.FileCount))
	fileCountLabel.TextStyle = fyne.TextStyle{Italic: true}

	cardContent := container.NewVBox(
		container.NewHBox(
			widget.NewIcon(riskIcon),
			targetCheck,
		),
		descLabel,
		fileCountLabel,
	)

	return widget.NewCard("", "", cardContent)
}

func updateCacheTotal(state *CacheCleanerState, totalLabel *canvas.Text, cleanBtn *widget.Button) {
	var selectedSize int64
	for _, t := range state.Targets {
		if state.SelectedTargets[t.ID] {
			selectedSize += t.TotalSize
		}
	}
	totalLabel.Text = fmt.Sprintf("Selected: %s", fsutil.FormatBytes(selectedSize))
	totalLabel.Refresh()

	if selectedSize > 0 {
		cleanBtn.Enable()
	} else {
		cleanBtn.Disable()
	}
}

func startCacheClean(state *CacheCleanerState) {
	// Calculate selected size
	var selectedSize int64
	var selectedTargets []*cleaner.CleanTarget
	for _, t := range state.Targets {
		if state.SelectedTargets[t.ID] {
			selectedSize += t.TotalSize
			selectedTargets = append(selectedTargets, t)
		}
	}

	if selectedSize == 0 {
		dialog.ShowInformation("Nothing to clean", "No targets selected for cleaning", state.Window)
		return
	}

	// Confirm
	dialog.ShowConfirm(
		"Confirm Cleaning",
		fmt.Sprintf("Delete %s of cache files?", fsutil.FormatBytes(selectedSize)),
		func(ok bool) {
			if !ok {
				return
			}

			state.IsCleaning = true

			// Show progress view
			progressLabel := widget.NewLabel("Cleaning...")
			progressBar := widget.NewProgressBar()
			progressBar.SetValue(0)

			progressBody := container.NewVBox(
				layout.NewSpacer(),
				widget.NewCard("Progress", "Cleaning selected caches", container.NewVBox(
					progressLabel,
					progressBar,
				)),
				layout.NewSpacer(),
			)
			state.updateContent(createToolPage("Cleaning", "Please wait while we clean your system", progressBody))

			go func() {
				cleaned := 0
				var cleanedBytes int64

				for i, t := range selectedTargets {
					// Clean each path for this target
					for _, path := range t.Paths {
						deleted, freed, _ := cleanPath(path, t.Patterns)
						cleaned += deleted
						cleanedBytes += freed
					}

					fyne.Do(func() {
						progressBar.SetValue(float64(i+1) / float64(len(selectedTargets)))
						progressLabel.SetText(fmt.Sprintf("Cleaned: %s", t.Label))
					})
				}

				state.CleanedCount = cleaned
				state.CleanedBytes = cleanedBytes
				state.IsCleaning = false

				log.Printf("[CacheCleaner] Total deleted: %d items, %s", cleaned, fsutil.FormatBytes(cleanedBytes))

				fyne.Do(func() {
					showCacheCleanComplete(state)
				})
			}()
		},
		state.Window,
	)
}

// isProtectedPath returns true for system-protected directories
func isProtectedPath(path string) bool {
	if path == "" {
		return false
	}

	// Normalize path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't resolve the path, be conservative and protect it
		return true
	}

	var protected []string
	switch runtime.GOOS {
	case "darwin":
		protected = []string{
			"/var/folders",
			"/private/var",
			"/System",
			"/Library/Caches/com.apple",
		}
	case "windows":
		// Windows system paths (common locations)
		protected = []string{
			`C:\Windows`,
			`C:\Program Files`,
			`C:\Program Files (x86)`,
			`C:\ProgramData`,
		}
	case "linux":
		protected = []string{
			"/etc",
			"/bin",
			"/usr",
			"/sbin",
			"/lib",
			"/boot",
			"/dev",
			"/proc",
			"/sys",
		}
	}

	for _, p := range protected {
		// Check if path matches or is within a protected path
		// We use strings.HasPrefix but need to ensure it's a full path match or a bundle ID subpath
		if absPath == p ||
			strings.HasPrefix(absPath, p+string(filepath.Separator)) ||
			strings.HasPrefix(absPath, p+"/") ||
			strings.HasPrefix(absPath, p+".") {
			return true
		}
	}
	return false
}

// cleanPath deletes files matching the pattern at the given path
func cleanPath(basePath string, patterns []string) (int, int64, error) {
	deleted := 0
	var freedBytes int64

	// Check if base path exists
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		log.Printf("[CacheCleaner] Path does not exist: %s", basePath)
		return 0, 0, nil
	}

	// For patterns like "*", just delete the entire directory contents recursively
	if len(patterns) == 1 && patterns[0] == "*" {
		// First, measure what we're deleting
		var measuredBytes int64
		var measuredCount int
		err := filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors, continue deleting
			}
			if path == basePath {
				return nil // Don't count the base directory itself
			}
			// Track what we're deleting
			info, err := d.Info()
			if err != nil {
				return nil
			}
			if !d.IsDir() {
				measuredBytes += info.Size()
			}
			measuredCount++
			return nil
		})
		if err != nil {
			log.Printf("[CacheCleaner] Error measuring %s: %v", basePath, err)
		}

		log.Printf("[CacheCleaner] Measured %s (%d files) in %s", fsutil.FormatBytes(measuredBytes), measuredCount, basePath)

		// Now actually delete everything
		if err := os.RemoveAll(basePath); err != nil {
			log.Printf("[CacheCleaner] Error deleting %s: %v", basePath, err)
			return measuredCount, measuredBytes, err
		}
		// Recreate the base directory
		if err := os.MkdirAll(basePath, 0755); err != nil {
			log.Printf("[CacheCleaner] Error recreating %s: %v", basePath, err)
			return measuredCount, measuredBytes, err
		}

		log.Printf("[CacheCleaner] Deleted %s from %s", fsutil.FormatBytes(measuredBytes), basePath)
		return measuredCount, measuredBytes, nil
	}

	// Get all entries in the path
	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	for _, entry := range entries {
		// Check if entry matches any pattern
		matched := false
		for _, pattern := range patterns {
			if matched, _ = filepath.Match(pattern, entry.Name()); matched {
				break
			}
		}

		if !matched {
			continue
		}

		fullPath := filepath.Join(basePath, entry.Name())

		// Get file info and size
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		// Delete the file/directory recursively
		if err := os.RemoveAll(fullPath); err != nil {
			// Try trash as fallback for single files
			_ = moveToTrash(fullPath)
		}

		deleted++
		freedBytes += info.Size()
	}

	log.Printf("[CacheCleaner] Deleted %s from %s (pattern match)", fsutil.FormatBytes(freedBytes), basePath)
	return deleted, freedBytes, nil
}

func showCacheCleanComplete(state *CacheCleanerState) {
	title := canvas.NewText("Cleaning Complete!", theme.Color(theme.ColorNameSuccess))
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	icon := canvas.NewImageFromResource(theme.ConfirmIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(80, 80))

	message := fmt.Sprintf("Cleaned %d cache locations", state.CleanedCount)
	subMessage := fmt.Sprintf("Freed %s of disk space", fsutil.FormatBytes(state.CleanedBytes))

	resultLabel := widget.NewLabel(message)
	resultLabel.TextStyle = fyne.TextStyle{Bold: true}
	resultLabel.Alignment = fyne.TextAlignCenter

	subLabel := widget.NewLabel(subMessage)
	subLabel.TextStyle = fyne.TextStyle{Italic: true}
	subLabel.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButtonWithIcon("Scan Again", theme.ViewRefreshIcon(), func() {
		state.updateContent(CacheCleanerWidget(state))
	})
	backBtn.Importance = widget.HighImportance

	body := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(container.NewVBox(
			icon,
			title,
			resultLabel,
			subLabel,
			layout.NewSpacer(),
			container.NewHBox(layout.NewSpacer(), backBtn, layout.NewSpacer()),
		)),
		layout.NewSpacer(),
	)

	state.updateContent(createToolPage("Cleaning Finished", "Summary of the cleaning operation", body))
}

// NewCacheCleanerState creates a new cache cleaner state
func NewCacheCleanerState(window fyne.Window) *CacheCleanerState {
	return &CacheCleanerState{
		Window:          window,
		SelectedTargets: make(map[string]bool),
		Concurrency:     4, // default
	}
}

// parseDuration wraps time.ParseDuration to support days ('d')
func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
