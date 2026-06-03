package gui

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/PopolQue/dupclean/cleaner"
	"github.com/PopolQue/dupclean/gui/components"
	"github.com/PopolQue/dupclean/internal/fsutil"

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
	mu                     sync.Mutex
	FailedCleanPaths       []string
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
	state.mu.Lock()
	state.IsScanning = true
	state.mu.Unlock()
	state.ProcessManager.SetProcessRunning(true)

	comp := state.cacheCleanerComponents
	comp.scanBtn.Disable()
	comp.minAgeEntry.Disable()

	comp.progressLabel.SetText("Scanning for caches...")
	comp.progressBar.Show()
	comp.progressBar.SetValue(0)

	go func() {
		result, err := performCacheScan(state)
		if err != nil {
			fyne.Do(func() {
				state.ProcessManager.SetProcessRunning(false)
				comp.progressLabel.SetText(fmt.Sprintf("Error: %v", err))
				comp.scanBtn.Enable()
				comp.minAgeEntry.Enable()
			})
			return
		}

		updateStateAfterScan(state, result)
		state.ProcessManager.SetProcessRunning(false)

		fyne.Do(func() {
			displayCacheResults(state)
		})
	}()
}

func updateStateAfterScan(state *CacheCleanerState, result *cleaner.ScanResult) {
	state.mu.Lock()
	defer state.mu.Unlock()
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
}

func performCacheScan(state *CacheCleanerState) (*cleaner.ScanResult, error) {
	// Parse min age
	minAge, err := fsutil.ParseDuration(state.MinAgeStr)
	if err != nil {
		return nil, err
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
	comp := state.cacheCleanerComponents
	opts := cleaner.ScanOptions{
		Concurrency: state.Concurrency,
		MinAge:      minAge,
		OnProgress: func(progress cleaner.Progress) {
			if comp != nil {
				fyne.Do(func() {
					if progress.Total > 0 {
						comp.progressBar.SetValue(float64(progress.Done) / float64(progress.Total))
					}
					comp.progressLabel.SetText(fmt.Sprintf("Scanning: %s", progress.Current))
				})
			}
		},
	}

	return cleaner.Scan(targets, opts)
}

func groupTargetsByCategory(targets []*cleaner.CleanTarget) map[string][]*cleaner.CleanTarget {
	categories := make(map[string][]*cleaner.CleanTarget)
	for _, t := range targets {
		if t.TotalSize > 0 {
			categories[t.Category] = append(categories[t.Category], t)
		}
	}
	return categories
}

func getSortedCategoryNames(categories map[string][]*cleaner.CleanTarget) []string {
	catNames := make([]string, 0, len(categories))
	for cat := range categories {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)
	return catNames
}

func displayCacheResults(state *CacheCleanerState) {
	// Accordion for categories
	accordion := widget.NewAccordion()

	// Group by category
	categories := groupTargetsByCategory(state.Targets)

	// Sort categories
	catNames := getSortedCategoryNames(categories)

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

func getRiskInfo(risk cleaner.Risk) (string, widget.Importance, fyne.Resource) {
	switch risk {
	case cleaner.RiskModerate:
		return "Moderate", widget.WarningImportance, theme.WarningIcon()
	case cleaner.RiskHigh:
		return "High Risk", widget.DangerImportance, theme.ErrorIcon()
	default:
		return "Safe", widget.SuccessImportance, theme.ConfirmIcon()
	}
}

func toggleTargetSelection(state *CacheCleanerState, targetID string, checked bool, totalLabel *widget.Label, cleanBtn *widget.Button) {
	state.SelectedTargets[targetID] = checked
	updateCacheTotal(state, totalLabel, cleanBtn)
}

// createCacheTargetCard creates a card for a cache target
func createCacheTargetCard(state *CacheCleanerState, target *cleaner.CleanTarget, totalLabel *widget.Label, cleanBtn *widget.Button) *widget.Card {
	// Risk icon and badge
	riskText, riskImportance, riskIcon := getRiskInfo(target.Risk)
	riskBadge := components.StatusBadge(riskText, riskImportance)

	// Target checkbox
	targetCheck := widget.NewCheck(fmt.Sprintf("%s (%s)", target.Label, fsutil.FormatBytes(target.TotalSize)), func(checked bool) {
		toggleTargetSelection(state, target.ID, checked, totalLabel, cleanBtn)
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

func calculateSelectedSize(state *CacheCleanerState) int64 {
	var selectedSize int64
	for _, t := range state.Targets {
		if state.SelectedTargets[t.ID] {
			selectedSize += t.TotalSize
		}
	}
	return selectedSize
}

func getUpdateCacheTotalState(state *CacheCleanerState) (string, bool) {
	selectedSize := calculateSelectedSize(state)
	text := fmt.Sprintf("Selected: %s", fsutil.FormatBytes(selectedSize))
	enabled := selectedSize > 0
	return text, enabled
}

func updateCacheTotal(state *CacheCleanerState, totalLabel *widget.Label, cleanBtn *widget.Button) {
	text, enabled := getUpdateCacheTotalState(state)
	totalLabel.SetText(text)

	if enabled {
		cleanBtn.Enable()
	} else {
		cleanBtn.Disable()
	}
}

func getSelectedTargets(state *CacheCleanerState) ([]*cleaner.CleanTarget, int64) {
	var selectedSize int64
	var selectedTargets []*cleaner.CleanTarget
	for _, t := range state.Targets {
		if state.SelectedTargets[t.ID] {
			selectedSize += t.TotalSize
			selectedTargets = append(selectedTargets, t)
		}
	}
	return selectedTargets, selectedSize
}

func startCacheClean(state *CacheCleanerState) {
	// Calculate selected size
	selectedTargets, selectedSize := getSelectedTargets(state)

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

			state.mu.Lock()
			state.IsCleaning = true
			state.mu.Unlock()

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
				// Perform actual cleaning synchronously
				cleaned, cleanedBytes, failedPaths := performCacheClean(selectedTargets, func(progress float64, currentLabel string) {
					fyne.Do(func() {
						progressBar.SetValue(progress)
						progressLabel.SetText(currentLabel)
					})
				})

				state.mu.Lock()
				state.CleanedCount = cleaned
				state.CleanedBytes = cleanedBytes
				state.FailedCleanPaths = failedPaths
				state.IsCleaning = false
				state.mu.Unlock()

				log.Printf("[CacheCleaner] Total deleted: %d items, %s", cleaned, fsutil.FormatBytes(cleanedBytes))

				fyne.Do(func() {
					showCacheCleanComplete(state)
				})
			}()
		},
		state.Window,
	)
}

// performCacheClean executes the cleaning operation synchronously.
// The onProgress callback allows updating UI or logging progress.
func performCacheClean(selectedTargets []*cleaner.CleanTarget, onProgress func(progress float64, currentLabel string)) (int, int64, []string) {
	cleaned := 0
	var cleanedBytes int64
	var failedPaths []string

	for i, t := range selectedTargets {
		// Clean each path for this target
		for _, path := range t.Paths {
			deleted, freed, err := cleanPath(path, t.Patterns)
			cleaned += deleted
			cleanedBytes += freed
			if err != nil {
				failedPaths = append(failedPaths, path)
			}
		}

		if onProgress != nil {
			onProgress(float64(i+1)/float64(len(selectedTargets)), fmt.Sprintf("Cleaned: %s", t.Label))
		}
	}

	return cleaned, cleanedBytes, failedPaths
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
		if err := moveToTrash(basePath); err != nil {
			log.Printf("[CacheCleaner] Fallback to permanent delete for %s: %v", basePath, err)
			_ = osRemoveAll(basePath)
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

		// Delete the file/directory recursively using trash API by default
		if err := moveToTrash(fullPath); err != nil {
			// Fallback to permanent delete if trash fails
			log.Printf("[CacheCleaner] Fallback to permanent delete for %s: %v", fullPath, err)
			_ = osRemoveAll(fullPath)
		}

		deleted++
		freedBytes += info.Size()
	}

	log.Printf("[CacheCleaner] Deleted %s from %s (pattern match)", fsutil.FormatBytes(freedBytes), basePath)
	return deleted, freedBytes, nil
}

func getCleanCompleteSummary(cleanedCount int, cleanedBytes int64, failedPaths []string) (string, string) {
	message := fmt.Sprintf("Cleaned %d cache locations", cleanedCount)
	subMessage := fmt.Sprintf("Freed %s of disk space", fsutil.FormatBytes(cleanedBytes))

	if len(failedPaths) > 0 {
		subMessage += fmt.Sprintf("\n⚠️ %d location(s) had errors or were in use", len(failedPaths))

		showMax := 3
		if len(failedPaths) < showMax {
			showMax = len(failedPaths)
		}
		subMessage += "\nFailed: " + strings.Join(failedPaths[:showMax], ", ")
		if len(failedPaths) > showMax {
			subMessage += fmt.Sprintf("... and %d more", len(failedPaths)-showMax)
		}
	}

	return message, subMessage
}

func showCacheCleanComplete(state *CacheCleanerState) {
	state.mu.Lock()
	cleanedCount := state.CleanedCount
	cleanedBytes := state.CleanedBytes
	failedPaths := state.FailedCleanPaths
	state.mu.Unlock()

	message, subMessage := getCleanCompleteSummary(cleanedCount, cleanedBytes, failedPaths)

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

// NewCacheCleanerState creates a new cache cleaner state.
// If concurrency is 0, it defaults to 4.
func NewCacheCleanerState(window fyne.Window, pm *ProcessManager, concurrency int) *CacheCleanerState {
	if concurrency == 0 {
		concurrency = 4
	}
	return &CacheCleanerState{
		Window:          window,
		ProcessManager:  pm,
		SelectedTargets: make(map[string]bool),
		Concurrency:     concurrency,
	}
}
