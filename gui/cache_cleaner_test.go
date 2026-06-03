package gui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/widget"
	"github.com/PopolQue/dupclean/cleaner"
)

func TestIsProtectedPath(t *testing.T) {
	oldOS := goos
	oldAbs := absPath
	oldSep := pathSeparator
	defer func() {
		goos = oldOS
		absPath = oldAbs
		pathSeparator = oldSep
	}()

	// Mock absPath to return the path as-is for testing
	absPath = func(path string) (string, error) { return path, nil }

	tests := []struct {
		path     string
		expected bool
		os       string
		sep      string
	}{
		// macOS paths
		{"/var/folders/abc123", true, "darwin", "/"},
		{"/var/folders", true, "darwin", "/"},
		{"/private/var", true, "darwin", "/"},
		{"/System/Library", true, "darwin", "/"},
		{"/Library/Caches/com.apple", true, "darwin", "/"},

		// Windows paths
		{`C:\Windows\System32`, true, "windows", `\`},
		{`C:\Program Files\Common`, true, "windows", `\`},

		// Linux paths
		{"/etc/passwd", true, "linux", "/"},
		{"/usr/bin", true, "linux", "/"},

		// Common non-protected paths
		{"/Users/user/Library/Caches", false, "darwin", "/"},
		{"/tmp", false, "darwin", "/"},
		{"/tmp", false, "linux", "/"},
		{`C:\Users\test\Downloads`, false, "windows", `\`},
		{"", false, "", "/"},
	}

	for _, test := range tests {
		t.Run(test.os+"_"+test.path, func(t *testing.T) {
			if test.os != "" {
				goos = test.os
			}
			if test.sep != "" {
				pathSeparator = test.sep
			}
			result := isProtectedPath(test.path)
			if result != test.expected {
				t.Errorf("isProtectedPath(%q) on %s (sep=%s) = %v, want %v", test.path, test.os, test.sep, result, test.expected)
			}
		})
	}
}

func TestCleanPath_NonExistent(t *testing.T) {
	deleted, freed, err := cleanPath("/nonexistent/path/that/does/not/exist", []string{"*"})

	if err != nil {
		t.Errorf("cleanPath() with non-existent path returned error: %v", err)
	}

	if deleted != 0 {
		t.Errorf("cleanPath() returned deleted = %d, expected 0", deleted)
	}

	if freed != 0 {
		t.Errorf("cleanPath() returned freed = %d, expected 0", freed)
	}
}

func TestCleanPath_StarPattern(t *testing.T) {
	oldTrash := moveToTrash
	moveToTrash = func(path string) error { return nil }
	defer func() { moveToTrash = oldTrash }()

	tmpDir := t.TempDir()

	// Create test files
	writeFile(t, tmpDir, "file1.tmp", "content1")
	writeFile(t, tmpDir, "file2.tmp", "content2")
	writeFile(t, tmpDir, "subdir/file3.tmp", "content3")

	deleted, freed, err := cleanPath(tmpDir, []string{"*"})

	if err != nil {
		t.Errorf("cleanPath() error = %v", err)
	}

	if deleted == 0 {
		t.Error("cleanPath() returned deleted = 0, expected > 0")
	}

	if freed == 0 {
		t.Error("cleanPath() returned freed = 0, expected > 0")
	}
}

func TestCleanPath_SpecificPattern(t *testing.T) {
	oldTrash := moveToTrash
	moveToTrash = func(path string) error { return nil }
	defer func() { moveToTrash = oldTrash }()

	tmpDir := t.TempDir()

	// Create test files with different extensions
	writeFile(t, tmpDir, "file1.log", "log content")
	writeFile(t, tmpDir, "file2.tmp", "temp content")
	writeFile(t, tmpDir, "file3.dat", "data content")

	// Only delete .log files
	deleted, freed, err := cleanPath(tmpDir, []string{"*.log"})

	if err != nil {
		t.Errorf("cleanPath() error = %v", err)
	}

	if deleted != 1 {
		t.Errorf("cleanPath() returned deleted = %d, expected 1", deleted)
	}

	if freed == 0 {
		t.Error("cleanPath() returned freed = 0, expected > 0")
	}
}

func TestCleanPath_EmptyPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, tmpDir, "file1.tmp", "content1")
	writeFile(t, tmpDir, "file2.tmp", "content2")

	// Empty patterns should match nothing
	deleted, freed, err := cleanPath(tmpDir, []string{})

	if err != nil {
		t.Errorf("cleanPath() error = %v", err)
	}

	if deleted != 0 {
		t.Errorf("cleanPath() with empty patterns returned deleted = %d, expected 0", deleted)
	}

	if freed != 0 {
		t.Errorf("cleanPath() with empty patterns returned freed = %d, expected 0", freed)
	}
}

func TestPerformCacheScan_Logic(t *testing.T) {
	// Create a minimal state
	state := &CacheCleanerState{
		MinAgeStr:   "1h",
		Concurrency: 1,
	}

	result, err := performCacheScan(state)
	if err != nil {
		t.Fatalf("performCacheScan() error = %v", err)
	}

	if result == nil {
		t.Fatal("performCacheScan() returned nil result")
	}
}

func TestUpdateStateAfterScan(t *testing.T) {
	state := &CacheCleanerState{}
	result := &cleaner.ScanResult{
		Targets: []*cleaner.CleanTarget{
			{ID: "safe", TotalSize: 100, Risk: cleaner.RiskSafe},
			{ID: "risky", TotalSize: 100, Risk: cleaner.RiskHigh},
		},
		TotalSize: 200,
	}

	updateStateAfterScan(state, result)

	if state.TotalSize != 200 {
		t.Errorf("TotalSize = %d, want 200", state.TotalSize)
	}
	if !state.SelectedTargets["safe"] {
		t.Error("Target 'safe' should be selected")
	}
	if state.SelectedTargets["risky"] {
		t.Error("Target 'risky' should not be selected")
	}
}

func TestGroupTargetsByCategory(t *testing.T) {
	targets := []*cleaner.CleanTarget{
		{ID: "1", Category: "Browser", TotalSize: 100},
		{ID: "2", Category: "System", TotalSize: 200},
		{ID: "3", Category: "Browser", TotalSize: 0}, // Should be filtered
	}

	grouped := groupTargetsByCategory(targets)
	if len(grouped) != 2 {
		t.Errorf("groupTargetsByCategory() length = %d, want 2", len(grouped))
	}
	if len(grouped["Browser"]) != 1 {
		t.Errorf("groupTargetsByCategory() Browser length = %d, want 1", len(grouped["Browser"]))
	}
}

func TestGetSortedCategoryNames(t *testing.T) {
	categories := map[string][]*cleaner.CleanTarget{
		"System":  {},
		"Browser": {},
		"App":     {},
	}

	names := getSortedCategoryNames(categories)
	expected := []string{"App", "Browser", "System"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("getSortedCategoryNames() at %d = %s, want %s", i, name, expected[i])
		}
	}
}

func TestGetRiskInfo(t *testing.T) {
	tests := []struct {
		risk           cleaner.Risk
		expectedText   string
		expectedImport widget.Importance
	}{
		{cleaner.RiskModerate, "Moderate", widget.WarningImportance},
		{cleaner.RiskHigh, "High Risk", widget.DangerImportance},
		{cleaner.RiskSafe, "Safe", widget.SuccessImportance},
	}

	for _, tc := range tests {
		text, importance, _ := getRiskInfo(tc.risk)
		if text != tc.expectedText {
			t.Errorf("getRiskInfo(%v) text = %s, want %s", tc.risk, text, tc.expectedText)
		}
		if importance != tc.expectedImport {
			t.Errorf("getRiskInfo(%v) importance = %v, want %v", tc.risk, importance, tc.expectedImport)
		}
	}
}

func TestToggleTargetSelection(t *testing.T) {
	state := &CacheCleanerState{
		Targets: []*cleaner.CleanTarget{
			{ID: "1", TotalSize: 100},
		},
		SelectedTargets: map[string]bool{"1": false},
	}
	// Using dummy widgets as the logic doesn't actually depend on their state,
	// only on the callbacks they trigger which we are mocking by calling the function directly.
	toggleTargetSelection(state, "1", true, widget.NewLabel(""), widget.NewButton("", nil))

	if !state.SelectedTargets["1"] {
		t.Error("toggleTargetSelection() should set SelectedTargets['1'] to true")
	}
}

func TestCalculateSelectedSize(t *testing.T) {
	state := &CacheCleanerState{
		Targets: []*cleaner.CleanTarget{
			{ID: "1", TotalSize: 100},
			{ID: "2", TotalSize: 200},
			{ID: "3", TotalSize: 300},
		},
		SelectedTargets: map[string]bool{
			"1": true,
			"3": true,
		},
	}

	size := calculateSelectedSize(state)
	if size != 400 {
		t.Errorf("calculateSelectedSize() = %d, want 400", size)
	}
}

func TestGetSelectedTargets(t *testing.T) {
	state := &CacheCleanerState{
		Targets: []*cleaner.CleanTarget{
			{ID: "1", TotalSize: 100},
			{ID: "2", TotalSize: 200},
			{ID: "3", TotalSize: 300},
		},
		SelectedTargets: map[string]bool{
			"1": true,
			"3": true,
		},
	}

	targets, size := getSelectedTargets(state)
	if size != 400 {
		t.Errorf("getSelectedTargets() size = %d, want 400", size)
	}
	if len(targets) != 2 {
		t.Errorf("getSelectedTargets() length = %d, want 2", len(targets))
	}
}

func TestPerformCacheClean_Logic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	writeFile(t, tmpDir, "file1.tmp", "content1")
	writeFile(t, tmpDir, "file2.tmp", "content2")

	targets := []*cleaner.CleanTarget{
		{
			Label:    "Test Target",
			Paths:    []string{tmpDir},
			Patterns: []string{"*.tmp"},
		},
	}

	progressCalled := false
	cleaned, cleanedBytes := performCacheClean(targets, func(progress float64, currentLabel string) {
		progressCalled = true
	})

	if cleaned == 0 {
		t.Error("performCacheClean() expected > 0 files cleaned")
	}
	if cleanedBytes == 0 {
		t.Error("performCacheClean() expected > 0 bytes cleaned")
	}
	if !progressCalled {
		t.Error("performCacheClean() expected progress callback to be called")
	}
}

func TestGetCleanCompleteSummary(t *testing.T) {
	message, subMessage := getCleanCompleteSummary(5, 1024*1024)
	if message != "Cleaned 5 cache locations" {
		t.Errorf("Unexpected message: %s", message)
	}
	if subMessage != "Freed 1.0 MB of disk space" {
		t.Errorf("Unexpected subMessage: %s", subMessage)
	}
}

func TestNewCacheCleanerState(t *testing.T) {
	// Test default concurrency
	state := NewCacheCleanerState(nil, nil, 0)
	if state.Concurrency != 4 {
		t.Errorf("NewCacheCleanerState default concurrency = %d, want 4", state.Concurrency)
	}

	// Test custom concurrency
	state2 := NewCacheCleanerState(nil, nil, 8)
	if state2.Concurrency != 8 {
		t.Errorf("NewCacheCleanerState custom concurrency = %d, want 8", state2.Concurrency)
	}

	if state.SelectedTargets == nil {
		t.Error("NewCacheCleanerState should initialize SelectedTargets")
	}
}

// Helper function to create test files
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()

	fullPath := filepath.Join(dir, name)

	// Create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
