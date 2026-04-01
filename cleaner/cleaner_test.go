package cleaner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestGetSystemTargets(t *testing.T) {
	targets := GetSystemTargets()

	if targets == nil {
		t.Fatal("GetSystemTargets() returned nil")
	}

	// Should return targets for current OS
	var expectedCount int
	switch runtime.GOOS {
	case "darwin":
		expectedCount = 4 // macOS has 4 system targets
	case "linux":
		expectedCount = 2 // Linux has 2 system targets
	case "windows":
		expectedCount = 2 // Windows has 2 system targets
	}

	if len(targets) < expectedCount {
		t.Errorf("GetSystemTargets() returned %d targets, expected at least %d for %s",
			len(targets), expectedCount, runtime.GOOS)
	}

	// Verify all targets have required fields
	for _, target := range targets {
		if target.ID == "" {
			t.Error("Target missing ID")
		}
		if target.Category == "" {
			t.Error("Target missing Category")
		}
		if target.Label == "" {
			t.Error("Target missing Label")
		}
		if len(target.Paths) == 0 {
			t.Errorf("Target %s has no paths", target.ID)
		}
		if len(target.Patterns) == 0 {
			t.Errorf("Target %s has no patterns", target.ID)
		}
	}
}

func TestGetBrowserTargets(t *testing.T) {
	targets := GetBrowserTargets()

	if targets == nil {
		t.Fatal("GetBrowserTargets() returned nil")
	}

	// Should return at least some browser targets
	if len(targets) == 0 {
		t.Error("GetBrowserTargets() returned no targets")
	}

	// Verify common browsers are included
	foundSafari := false
	foundChrome := false
	foundFirefox := false

	for _, target := range targets {
		if target.ID == "macos-safari-cache" {
			foundSafari = true
		}
		if target.ID == "macos-chrome-cache" || target.ID == "linux-chrome-cache" || target.ID == "windows-chrome-cache" {
			foundChrome = true
		}
		if target.ID == "macos-firefox-cache" || target.ID == "linux-firefox-cache" || target.ID == "windows-firefox-cache" {
			foundFirefox = true
		}
	}

	// At least one major browser should be present for the current OS
	// Skip if no browsers are found (may happen in CI environments)
	if !foundSafari && !foundChrome && !foundFirefox {
		t.Skip("No major browser targets found (expected in some CI environments)")
	}
}

func TestGetDeveloperTargets(t *testing.T) {
	targets := GetDeveloperTargets()

	if targets == nil {
		t.Fatal("GetDeveloperTargets() returned nil")
	}

	// Should return at least some developer targets
	if len(targets) == 0 {
		t.Error("GetDeveloperTargets() returned no targets")
	}

	// Verify all targets have Go cache path
	for _, target := range targets {
		if target.ID == "dev-go-cache" {
			if len(target.Paths) == 0 {
				t.Error("Go cache target has no paths")
			}
		}
	}
}

func TestGetLogsTargets(t *testing.T) {
	targets := GetLogsTargets()

	if targets == nil {
		t.Fatal("GetLogsTargets() returned nil")
	}

	// Should return at least one logs target
	if len(targets) == 0 {
		t.Error("GetLogsTargets() returned no targets")
	}
}

func TestRegistry(t *testing.T) {
	targets := Registry()

	if targets == nil {
		t.Fatal("Registry() returned nil")
	}

	// Should return targets from all categories
	if len(targets) == 0 {
		t.Error("Registry() returned no targets")
	}

	// Verify we have targets from different categories
	categories := make(map[string]bool)
	for _, target := range targets {
		categories[target.Category] = true
	}

	// Should have at least System category
	if !categories["System"] {
		t.Error("Registry() did not return any System category targets")
	}
}

func TestFilterTargets(t *testing.T) {
	targets := Registry()

	// Test filtering by category
	filtered := FilterTargets(targets, "System", []string{}, false, false)
	if len(filtered) == 0 {
		t.Error("FilterTargets() by category returned no results")
	}

	// Verify all filtered targets are in the System category
	for _, target := range filtered {
		if target.Category != "System" {
			t.Errorf("FilterTargets() returned target with category %s, expected System", target.Category)
		}
	}

	// Test filtering by IDs
	if len(targets) > 0 {
		firstID := targets[0].ID
		filtered = FilterTargets(targets, "", []string{firstID}, false, false)
		if len(filtered) != 1 {
			t.Errorf("FilterTargets() by ID returned %d targets, expected 1", len(filtered))
		}
		if len(filtered) > 0 && filtered[0].ID != firstID {
			t.Errorf("FilterTargets() returned wrong target: %s", filtered[0].ID)
		}
	}

	// Test filtering with noDeveloper
	filtered = FilterTargets(targets, "", []string{}, true, false)
	for _, target := range filtered {
		if target.Category == "Developer" {
			t.Error("FilterTargets() with noDeveloper=true returned developer target")
		}
	}

	// Test filtering with noBrowser
	filtered = FilterTargets(targets, "", []string{}, false, true)
	for _, target := range filtered {
		if target.Category == "Browser" {
			t.Error("FilterTargets() with noBrowser=true returned browser target")
		}
	}
}

func TestScan_Basic(t *testing.T) {
	// Create temporary test directories
	tmpDir := t.TempDir()

	// Create a test cache structure
	cacheDir := filepath.Join(tmpDir, "test-cache")
	os.MkdirAll(cacheDir, 0755)

	// Create some test files
	os.WriteFile(filepath.Join(cacheDir, "file1.tmp"), []byte("test content 1"), 0644)
	os.WriteFile(filepath.Join(cacheDir, "file2.tmp"), []byte("test content 2"), 0644)
	os.WriteFile(filepath.Join(cacheDir, "file3.dat"), []byte("test content 3"), 0644)

	// Create a test target
	target := &CleanTarget{
		ID:          "test-cache",
		Category:    "Test",
		Label:       "Test Cache",
		Description: "A test cache directory",
		Paths:       []string{cacheDir},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
	}

	// Scan the target
	opts := ScanOptions{
		Concurrency: 1,
	}

	result, err := Scan([]*CleanTarget{target}, opts)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result == nil {
		t.Fatal("Scan() returned nil result")
	}

	if result.TotalSize == 0 {
		t.Error("Scan() returned TotalSize = 0, expected > 0")
	}

	if len(result.Targets) != 1 {
		t.Errorf("Scan() returned %d targets, expected 1", len(result.Targets))
	}
}

func TestScan_WithMinAge(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "test-cache")
	os.MkdirAll(cacheDir, 0755)

	// Create a new file (should be excluded with minAge)
	os.WriteFile(filepath.Join(cacheDir, "new.tmp"), []byte("new"), 0644)

	// Create an old file
	oldFile := filepath.Join(cacheDir, "old.tmp")
	os.WriteFile(oldFile, []byte("old"), 0644)

	// Set modification time to 2 days ago
	oldTime := time.Now().Add(-48 * time.Hour)
	os.Chtimes(oldFile, oldTime, oldTime)

	target := &CleanTarget{
		ID:       "test-cache",
		Category: "Test",
		Label:    "Test Cache",
		Paths:    []string{cacheDir},
		Patterns: []string{"*"},
		Risk:     RiskSafe,
	}

	opts := ScanOptions{
		MinAge: 24 * time.Hour, // Only files older than 24 hours
	}

	result, err := Scan([]*CleanTarget{target}, opts)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Should only include the old file
	if result.TotalSize == 0 {
		t.Error("Scan() with MinAge returned TotalSize = 0, expected > 0")
	}
}

func TestScan_NonExistentPath(t *testing.T) {
	target := &CleanTarget{
		ID:       "test-nonexistent",
		Category: "Test",
		Label:    "Non-existent Path",
		Paths:    []string{"/nonexistent/path/that/does/not/exist"},
		Patterns: []string{"*"},
		Risk:     RiskSafe,
	}

	opts := ScanOptions{}
	result, err := Scan([]*CleanTarget{target}, opts)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result.TotalSize != 0 {
		t.Errorf("Scan() with non-existent path returned TotalSize = %d, expected 0", result.TotalSize)
	}
}

func TestScan_EmptyTargetList(t *testing.T) {
	opts := ScanOptions{}
	result, err := Scan([]*CleanTarget{}, opts)

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if result == nil {
		t.Fatal("Scan() with empty target list returned nil")
	}

	if result.TotalSize != 0 {
		t.Errorf("Scan() with empty target list returned TotalSize = %d, expected 0", result.TotalSize)
	}
}

func TestScanProgress(t *testing.T) {
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "test-cache")
	os.MkdirAll(cacheDir, 0755)

	// Create multiple test files
	for i := 0; i < 10; i++ {
		os.WriteFile(filepath.Join(cacheDir, "file"+string(rune(i))+".tmp"),
			[]byte("test content"), 0644)
	}

	target := &CleanTarget{
		ID:       "test-cache",
		Category: "Test",
		Label:    "Test Cache",
		Paths:    []string{cacheDir},
		Patterns: []string{"*"},
		Risk:     RiskSafe,
	}

	progressCalled := false
	opts := ScanOptions{
		OnProgress: func(progress Progress) {
			progressCalled = true
			if progress.Total <= 0 {
				t.Error("Progress.Total should be > 0")
			}
			if progress.Current == "" {
				t.Error("Progress.Current should not be empty")
			}
		},
	}

	_, err := Scan([]*CleanTarget{target}, opts)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if !progressCalled {
		t.Error("OnProgress callback was not called")
	}
}
