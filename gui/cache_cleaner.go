package gui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"dupclean/cleaner"

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
	cacheCleanerComponents *cacheCleanerComponents
}

type cacheCleanerComponents struct {
	scanBtn       *widget.Button
	results       *fyne.Container
	scroll        *container.Scroll
	cleanBtn      *widget.Button
	progressLabel *widget.Label
	progressBar   *widget.ProgressBar
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
	// Header
	title := canvas.NewText("Cache Cleaner", theme.Color(theme.ColorNamePrimary))
	title.TextSize = 28
	title.TextStyle = fyne.TextStyle{Bold: true}

	subtitle := canvas.NewText("Clean system, browser, and application caches", theme.Color(theme.ColorNameForeground))
	subtitle.TextSize = 14
	subtitle.TextStyle = fyne.TextStyle{Italic: true}

	// Disclaimer
	disclaimer := canvas.NewText("⚠️ Actual freed space may vary - cache files change constantly", theme.Color(theme.ColorNameWarning))
	disclaimer.TextSize = 12
	disclaimer.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(title, subtitle, disclaimer)

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

	// Results container
	resultsContainer := container.NewVBox()
	scroll := container.NewScroll(resultsContainer)
	scroll.SetMinSize(fyne.NewSize(700, 400))
	scroll.Hide()

	// Action buttons
	cleanBtn := widget.NewButtonWithIcon("Clean Selected", theme.ConfirmIcon(), func() {
		startCacheClean(state, resultsContainer, scroll, progressLabel, progressBar)
	})
	cleanBtn.Importance = widget.HighImportance
	cleanBtn.Disable()

	cancelBtn := widget.NewButton("Cancel", func() {
		state.Window.SetContent(CacheCleanerWidget(state))
	})
	cancelBtn.Importance = widget.LowImportance

	buttonRow := container.NewHBox(layout.NewSpacer(), cancelBtn, cleanBtn, layout.NewSpacer())

	content := container.NewVBox(
		header,
		layout.NewSpacer(),
		widget.NewCard("", "", container.NewVBox(scanBtn)),
		progressCard,
		scroll,
		layout.NewSpacer(),
		buttonRow,
		layout.NewSpacer(),
	)

	// Store references for updates
	state.cacheCleanerComponents = &cacheCleanerComponents{
		scanBtn:       scanBtn,
		results:       resultsContainer,
		scroll:        scroll,
		cleanBtn:      cleanBtn,
		progressLabel: progressLabel,
		progressBar:   progressBar,
	}

	return container.NewCenter(content)
}

func startCacheScan(state *CacheCleanerState) {
	state.IsScanning = true

	comp := state.cacheCleanerComponents
	comp.scanBtn.Disable()
	comp.progressLabel.SetText("Scanning for caches...")
	comp.progressBar.Show()
	comp.progressBar.SetValue(0)

	go func() {
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
			Concurrency: 4,
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
			comp.progressLabel.SetText(fmt.Sprintf("Found %s of cleanable caches", formatBytes(state.TotalSize)))
			comp.progressBar.SetValue(1.0)
			log.Printf("[CacheCleaner] Total cleanable: %s across %d targets", formatBytes(state.TotalSize), len(state.Targets))
			displayCacheResults(state, comp.results, comp.scroll, comp.cleanBtn)
		})
	}()
}

func displayCacheResults(state *CacheCleanerState, resultsContainer *fyne.Container, scroll *container.Scroll, cleanBtn *widget.Button) {
	resultsContainer.Objects = nil

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

	// Create total label - will be updated on selection
	totalLabel := canvas.NewText("Selected: 0 B", theme.Color(theme.ColorNamePrimary))

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

	// Add total label
	totalLabel.TextSize = 18
	totalLabel.TextStyle = fyne.TextStyle{Bold: true}
	resultsContainer.Add(totalLabel)

	scroll.Show()
	scroll.Refresh()
	cleanBtn.Disable() // Disable until something is selected
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
	targetCheck := widget.NewCheck(fmt.Sprintf("%s (%s)", target.Label, formatBytes(target.TotalSize)), func(checked bool) {
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
	totalLabel.Text = fmt.Sprintf("Selected: %s", formatBytes(selectedSize))
	totalLabel.Refresh()

	if selectedSize > 0 {
		cleanBtn.Enable()
	} else {
		cleanBtn.Disable()
	}
}

func startCacheClean(state *CacheCleanerState, resultsContainer *fyne.Container, scroll *container.Scroll, progressLabel *widget.Label, progressBar *widget.ProgressBar) {
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
		fmt.Sprintf("Delete %s of cache files?", formatBytes(selectedSize)),
		func(ok bool) {
			if !ok {
				return
			}

			state.IsCleaning = true
			progressLabel.SetText("Cleaning...")
			progressBar.SetValue(0)

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

				log.Printf("[CacheCleaner] Total deleted: %d items, %s", cleaned, formatBytes(cleanedBytes))
				log.Printf("[CacheCleaner] Expected to delete: %s", formatBytes(selectedSize))

				fyne.Do(func() {
					showCacheCleanComplete(state, progressLabel, progressBar, resultsContainer, scroll)
				})
			}()
		},
		state.Window,
	)
}

// isProtectedPath returns true for system-protected directories
func isProtectedPath(path string) bool {
	// macOS system paths
	protectedDarwin := []string{
		"/var/folders",
		"/private/var",
		"/System",
		"/Library/Caches/com.apple",
	}

	// Windows system paths
	protectedWindows := []string{
		`C:\Windows`,
		`C:\Program Files`,
		`C:\Program Files (x86)`,
		`C:\ProgramData`,
	}

	// Linux system paths
	protectedLinux := []string{
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

	var protected []string
	switch runtimeOS() {
	case "darwin":
		protected = protectedDarwin
	case "windows":
		protected = protectedWindows
	case "linux":
		protected = protectedLinux
	default:
		// Fallback: use all protected paths for unknown OS
		protected = append(protected, protectedDarwin...)
		protected = append(protected, protectedWindows...)
		protected = append(protected, protectedLinux...)
	}

	// Normalize path for comparison
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't resolve the path, be conservative and protect it
		return true
	}

	for _, p := range protected {
		// Check if path starts with protected path (with separator or exact match)
		if absPath == p || strings.HasPrefix(absPath, p+string(filepath.Separator)) {
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
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors, continue deleting
			}
			if path == basePath {
				return nil // Don't count the base directory itself
			}
			// Track what we're deleting
			if !info.IsDir() {
				measuredBytes += info.Size()
			}
			measuredCount++
			return nil
		})
		if err != nil {
			log.Printf("[CacheCleaner] Error measuring %s: %v", basePath, err)
		}

		log.Printf("[CacheCleaner] Measured %s (%d files) in %s", formatBytes(measuredBytes), measuredCount, basePath)

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

		log.Printf("[CacheCleaner] Deleted %s from %s", formatBytes(measuredBytes), basePath)
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

	log.Printf("[CacheCleaner] Deleted %s from %s (pattern match)", formatBytes(freedBytes), basePath)
	return deleted, freedBytes, nil
}

func showCacheCleanComplete(state *CacheCleanerState, progressLabel *widget.Label, progressBar *widget.ProgressBar, resultsContainer *fyne.Container, scroll *container.Scroll) {
	resultsContainer.Objects = nil

	title := canvas.NewText("Cleaning Complete!", theme.Color(theme.ColorNameSuccess))
	title.TextSize = 32
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	icon := canvas.NewImageFromResource(theme.ConfirmIcon())
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(80, 80))

	message := fmt.Sprintf("Cleaned %d cache locations", state.CleanedCount)
	subMessage := fmt.Sprintf("Freed %s of disk space", formatBytes(state.CleanedBytes))

	resultLabel := widget.NewLabel(message)
	resultLabel.TextStyle = fyne.TextStyle{Bold: true}
	resultLabel.Alignment = fyne.TextAlignCenter

	subLabel := widget.NewLabel(subMessage)
	subLabel.TextStyle = fyne.TextStyle{Italic: true}
	subLabel.Alignment = fyne.TextAlignCenter

	backBtn := widget.NewButtonWithIcon("Scan Again", theme.ViewRefreshIcon(), func() {
		// Refresh the cache cleaner widget
		state.updateContent(CacheCleanerWidget(state))
	})
	backBtn.Importance = widget.HighImportance

	content := container.NewVBox(
		icon,
		title,
		resultLabel,
		subLabel,
		layout.NewSpacer(),
		container.NewHBox(layout.NewSpacer(), backBtn, layout.NewSpacer()),
		layout.NewSpacer(),
	)

	scroll.Content = container.NewCenter(content)
	scroll.Refresh()
}

// NewCacheCleanerState creates a new cache cleaner state
func NewCacheCleanerState(window fyne.Window) *CacheCleanerState {
	return &CacheCleanerState{
		Window:          window,
		SelectedTargets: make(map[string]bool),
	}
}
