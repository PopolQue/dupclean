package gui

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"dupclean/cleaner"
	"dupclean/gui/components"
	"dupclean/internal/fsutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CacheCleanerState holds the state for the cache cleaner widget
type CacheCleanerState struct {
	Window                 fyne.Window
	ProcessManager         *ProcessManager
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

	// Performance options
	workerSlider := widget.NewSlider(1, 16)
	workerSlider.SetValue(float64(state.Concurrency))
	workerLabel := widget.NewLabel(fmt.Sprintf("%d", state.Concurrency))
	workerSlider.OnChanged = func(v float64) {
		val := int(v)
		state.Concurrency = val
		workerLabel.SetText(fmt.Sprintf("%d", val))
	}
	workerContainer := container.NewBorder(nil, nil, nil, workerLabel, workerSlider)

	optionsForm := widget.NewForm(
		widget.NewFormItem("Min Age", minAgeEntry),
		widget.NewFormItem("Workers", workerContainer),
	)
	optionsCard := widget.NewCard("Scan Settings", "Define filter criteria and performance", optionsForm)

	// Action
	scanBtn := widget.NewButtonWithIcon("Scan for Caches", theme.SearchIcon(), func() {
		startCacheScan(state)
	})
	scanBtn.Importance = widget.HighImportance
	state.ProcessManager.RegisterStartButton(scanBtn)

	progressBar := widget.NewProgressBar()
	progressBar.Hide()

	// Log/Menu
	disclaimer := widget.NewLabelWithStyle("⚠️ Actual freed space may vary - cache files change constantly", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	disclaimer.Importance = widget.WarningImportance
	logArea := container.NewVBox(disclaimer)

	// Store references for updates
	state.cacheCleanerComponents = &cacheCleanerComponents{
		scanBtn:       scanBtn,
		progressLabel: widget.NewLabel(""),
		progressBar:   progressBar,
		minAgeEntry:   minAgeEntry,
	}

	return components.FixedTabLayout(
		"Cache Cleaner",
		"Clean system, browser, and application caches",
		optionsCard,
		scanBtn,
		progressBar,
		logArea,
	)
}

func startCacheScan(state *CacheCleanerState) {
	state.IsScanning = true
	state.ProcessManager.SetProcessRunning(true)

	comp := state.cacheCleanerComponents
	comp.scanBtn.Disable()
	comp.minAgeEntry.Disable()

	comp.progressLabel.SetText("Scanning for caches...")
	comp.progressBar.Show()
	comp.progressBar.SetValue(0)

	go func() {
		// Parse min age
		minAge, err := fsutil.ParseDuration(state.MinAgeStr)
		if err != nil {
			fyne.Do(func() {
				state.ProcessManager.SetProcessRunning(false)
				comp.progressLabel.SetText(fmt.Sprintf("Invalid Min Age: %v", err))
				comp.scanBtn.Enable()
				comp.minAgeEntry.Enable()
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
				state.ProcessManager.SetProcessRunning(false)
				comp.progressLabel.SetText(fmt.Sprintf("Error: %v", err))
				comp.scanBtn.Enable()
				comp.minAgeEntry.Enable()
			})
			return
		}

		state.Targets = result.Targets
		state.TotalSize = result.TotalSize
		state.IsScanning = false
		state.ProcessManager.SetProcessRunning(false)

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
	// Accordion for categories
	accordion := widget.NewAccordion()

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
	totalLabel := widget.NewLabelWithStyle("Selected: 0 B", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	totalLabel.Importance = widget.HighImportance
	totalLabel.SizeName = theme.SizeNameSubHeadingText

	// Create accordion items
	for _, cat := range catNames {
		targets := categories[cat]
		catContent := container.NewVBox()

		for _, t := range targets {
			targetCard := createCacheTargetCard(state, t, totalLabel, cleanBtn)
			catContent.Add(targetCard)
		}

		accordion.Append(widget.NewAccordionItem(cat, catContent))
	}

	updateCacheTotal(state, totalLabel, cleanBtn)

	cancelBtn := widget.NewButton("Cancel", func() {
		state.updateContent(CacheCleanerWidget(state))
	})
	cancelBtn.Importance = widget.LowImportance

	buttonRow := components.ActionFooter(cancelBtn, totalLabel, cleanBtn)

	subtitle := fmt.Sprintf("Found %s of cleanable caches", fsutil.FormatBytes(state.TotalSize))
	state.updateContent(components.ToolPageWithFooter("Scan Results", subtitle, accordion, buttonRow))
}

// createCacheTargetCard creates a card for a cache target
func createCacheTargetCard(state *CacheCleanerState, target *cleaner.CleanTarget, totalLabel *widget.Label, cleanBtn *widget.Button) *widget.Card {
	// Risk icon and badge
	riskIcon := theme.ConfirmIcon()
	riskText := "Safe"
	riskImportance := widget.SuccessImportance
	switch target.Risk {
	case cleaner.RiskModerate:
		riskIcon = theme.WarningIcon()
		riskText = "Moderate"
		riskImportance = widget.WarningImportance
	case cleaner.RiskHigh:
		riskIcon = theme.ErrorIcon()
		riskText = "High Risk"
		riskImportance = widget.DangerImportance
	}
	riskBadge := components.StatusBadge(riskText, riskImportance)

	// Target checkbox
	targetCheck := widget.NewCheck(fmt.Sprintf("%s (%s)", target.Label, fsutil.FormatBytes(target.TotalSize)), func(checked bool) {
		state.SelectedTargets[target.ID] = checked
		updateCacheTotal(state, totalLabel, cleanBtn)
	})
	targetCheck.SetChecked(state.SelectedTargets[target.ID])

	// Description
	descLabel := widget.NewLabel(target.Description)
	descLabel.TextStyle = fyne.TextStyle{Italic: true}

	// Metadata
	metadata := fmt.Sprintf("%d items", target.FileCount)

	actions := container.NewHBox(riskBadge, widget.NewIcon(riskIcon))
	return components.ResultCard("", descLabel, metadata, targetCheck, actions)
}

func updateCacheTotal(state *CacheCleanerState, totalLabel *widget.Label, cleanBtn *widget.Button) {
	var selectedSize int64
	for _, t := range state.Targets {
		if state.SelectedTargets[t.ID] {
			selectedSize += t.TotalSize
		}
	}
	totalLabel.SetText(fmt.Sprintf("Selected: %s", fsutil.FormatBytes(selectedSize)))

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

			state.updateContent(components.ProgressPage(
				"Cleaning",
				"Please wait while we clean your system",
				"Progress",
				"Cleaning selected caches",
				progressLabel,
				progressBar,
			))

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
	abs, err := absPath(path)
	if err != nil {
		// If we can't resolve the path, be conservative and protect it
		return true
	}

	var protected []string
	switch goos {
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

	// Also protect the home directory itself
	if home, err := userHomeDir(); err == nil {
		if abs == home {
			return true
		}
	}

	for _, p := range protected {
		// Check if path matches or is within a protected path
		// We use strings.HasPrefix but need to ensure it's a full path match or a bundle ID subpath
		if abs == p ||
			strings.HasPrefix(abs, p+pathSeparator) ||
			strings.HasPrefix(abs, p+"/") ||
			strings.HasPrefix(abs, p+".") {
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
		if err := osRemoveAll(basePath); err != nil {
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
		if err := osRemoveAll(fullPath); err != nil {
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
	message := fmt.Sprintf("Cleaned %d cache locations", state.CleanedCount)
	subMessage := fmt.Sprintf("Freed %s of disk space", fsutil.FormatBytes(state.CleanedBytes))

	backBtn := widget.NewButtonWithIcon("Scan Again", theme.ViewRefreshIcon(), func() {
		state.updateContent(CacheCleanerWidget(state))
	})
	backBtn.Importance = widget.HighImportance

	state.updateContent(components.StatusPage(
		"Cleaning Finished",
		"Summary of the cleaning operation",
		theme.ConfirmIcon(),
		"Cleaning Complete!",
		message+"\n"+subMessage,
		backBtn,
	))
}

// NewCacheCleanerState creates a new cache cleaner state
func NewCacheCleanerState(window fyne.Window, pm *ProcessManager) *CacheCleanerState {
	return &CacheCleanerState{
		Window:          window,
		ProcessManager:  pm,
		SelectedTargets: make(map[string]bool),
		Concurrency:     4, // default
	}
}
