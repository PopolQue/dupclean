package gui

import (
	"os"
	"testing"

	"dupclean/cleaner"
)

// Test CacheCleanerState initialization
func TestCacheCleanerState_Init(t *testing.T) {
	state := &CacheCleanerState{
		Targets:         make([]*cleaner.CleanTarget, 0),
		SelectedTargets: make(map[string]bool),
		TotalSize:       0,
		IsScanning:      false,
		IsCleaning:      false,
		CleanedCount:    0,
		CleanedBytes:    0,
	}

	if state.Targets == nil {
		t.Error("Targets should be initialized")
	}
	if state.SelectedTargets == nil {
		t.Error("SelectedTargets should be initialized")
	}
	if state.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0", state.TotalSize)
	}
	if state.IsScanning {
		t.Error("IsScanning should be false")
	}
	if state.IsCleaning {
		t.Error("IsCleaning should be false")
	}
}

// Test CacheCleanerState with targets
func TestCacheCleanerState_WithTargets(t *testing.T) {
	state := &CacheCleanerState{
		Targets: []*cleaner.CleanTarget{
			{
				ID:       "test-target-1",
				Category: "Test",
				Label:    "Test Target 1",
			},
			{
				ID:       "test-target-2",
				Category: "Test",
				Label:    "Test Target 2",
			},
		},
		SelectedTargets: map[string]bool{
			"test-target-1": true,
		},
		TotalSize:    1024 * 1024,
		IsScanning:   false,
		IsCleaning:   false,
		CleanedCount: 0,
		CleanedBytes: 0,
	}

	if len(state.Targets) != 2 {
		t.Errorf("Targets length = %d, want 2", len(state.Targets))
	}
	if len(state.SelectedTargets) != 1 {
		t.Errorf("SelectedTargets length = %d, want 1", len(state.SelectedTargets))
	}
	if !state.SelectedTargets["test-target-1"] {
		t.Error("test-target-1 should be selected")
	}
	if state.SelectedTargets["test-target-2"] {
		t.Error("test-target-2 should not be selected")
	}
}

// Test updateContent with nil ContentContainer
func TestCacheCleanerState_UpdateContent_Nil(t *testing.T) {
	state := &CacheCleanerState{
		ContentContainer: nil,
	}

	// Should not panic
	state.updateContent(nil)
}

// Test cacheCleanerComponents initialization
func TestCacheCleanerComponents_Init(t *testing.T) {
	components := &cacheCleanerComponents{
		scanBtn:       nil,
		results:       nil,
		scroll:        nil,
		cleanBtn:      nil,
		progressLabel: nil,
		progressBar:   nil,
	}

	if components.scanBtn != nil {
		t.Error("scanBtn should be nil initially")
	}
	if components.results != nil {
		t.Error("results should be nil initially")
	}
	if components.scroll != nil {
		t.Error("scroll should be nil initially")
	}
	if components.cleanBtn != nil {
		t.Error("cleanBtn should be nil initially")
	}
	if components.progressLabel != nil {
		t.Error("progressLabel should be nil initially")
	}
	if components.progressBar != nil {
		t.Error("progressBar should be nil initially")
	}
}

// Test CacheCleanerWidget returns valid result
func TestCacheCleanerWidget_ReturnsValid(t *testing.T) {
	state := &CacheCleanerState{
		Targets:         make([]*cleaner.CleanTarget, 0),
		SelectedTargets: make(map[string]bool),
	}

	widget := CacheCleanerWidget(state)
	if widget == nil {
		t.Fatal("CacheCleanerWidget should return a non-nil widget")
	}
}

// Test CacheCleanerWidget with selected targets
func TestCacheCleanerWidget_WithSelectedTargets(t *testing.T) {
	state := &CacheCleanerState{
		Targets: []*cleaner.CleanTarget{
			{
				ID:       "target-1",
				Category: "Browser",
				Label:    "Chrome Cache",
			},
		},
		SelectedTargets: map[string]bool{
			"target-1": true,
		},
		TotalSize: 1024 * 1024 * 100, // 100 MB
	}

	widget := CacheCleanerWidget(state)
	if widget == nil {
		t.Fatal("CacheCleanerWidget should return a non-nil widget")
	}
}

// Test CacheCleanerWidget with large total size
func TestCacheCleanerWidget_LargeTotalSize(t *testing.T) {
	state := &CacheCleanerState{
		Targets: []*cleaner.CleanTarget{
			{
				ID:       "target-1",
				Category: "System",
				Label:    "Temp Files",
			},
			{
				ID:       "target-2",
				Category: "System",
				Label:    "Cache Files",
			},
		},
		SelectedTargets: map[string]bool{
			"target-1": true,
			"target-2": true,
		},
		TotalSize: 1024 * 1024 * 1024 * 5, // 5 GB
	}

	widget := CacheCleanerWidget(state)
	if widget == nil {
		t.Fatal("CacheCleanerWidget should return a non-nil widget")
	}
}

// Test CacheCleanerWidget with empty targets
func TestCacheCleanerWidget_EmptyTargets(t *testing.T) {
	state := &CacheCleanerState{
		Targets:         []*cleaner.CleanTarget{},
		SelectedTargets: make(map[string]bool),
		TotalSize:       0,
	}

	widget := CacheCleanerWidget(state)
	if widget == nil {
		t.Fatal("CacheCleanerWidget should return a non-nil widget")
	}
}

// Test CacheCleanerWidget with nil targets
func TestCacheCleanerWidget_NilTargets(t *testing.T) {
	state := &CacheCleanerState{
		Targets:         nil,
		SelectedTargets: make(map[string]bool),
		TotalSize:       0,
	}

	widget := CacheCleanerWidget(state)
	if widget == nil {
		t.Fatal("CacheCleanerWidget should return a non-nil widget")
	}
}

// Test isProtectedPath function
func TestIsProtectedPath_ProtectedPaths(t *testing.T) {
	// Test paths that should be protected based on actual implementation
	protectedPaths := []string{
		"/var/folders",
		"/var/folders/test",
		"/private/var",
		"/private/var/test",
		"/System",
		"/System/test",
		"/Library/Caches/com.apple",
		"/Library/Caches/com.apple.test",
	}

	for _, path := range protectedPaths {
		if !isProtectedPath(path) {
			t.Errorf("isProtectedPath(%q) should return true", path)
		}
	}
}

// Test isProtectedPath with non-protected paths
func TestIsProtectedPath_NonProtectedPaths(t *testing.T) {
	nonProtectedPaths := []string{
		"/Users/test/Cache",
		"/home/user/.cache",
		"C:\\Users\\test\\AppData\\Local\\Temp",
		"/tmp",
		"/var/tmp",
	}

	for _, path := range nonProtectedPaths {
		if isProtectedPath(path) {
			t.Errorf("isProtectedPath(%q) should return false", path)
		}
	}
}

// Test isProtectedPath with empty path
func TestIsProtectedPath_EmptyPath(t *testing.T) {
	if isProtectedPath("") {
		t.Error("isProtectedPath(\"\") should return false")
	}
}

// Test isProtectedPath case sensitivity
func TestIsProtectedPath_CaseSensitivity(t *testing.T) {
	// macOS paths are case-insensitive for some paths
	if !isProtectedPath("/system") {
		t.Log("isProtectedPath may be case-sensitive")
	}
}

// Test cleanPath with valid path
func TestCleanPath_ValidPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := tmpDir + "/test.txt"
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Clean with wildcard pattern
	count, bytes, err := cleanPath(tmpDir, []string{"*"})
	if err != nil {
		t.Errorf("cleanPath returned error: %v", err)
	}
	if count < 1 {
		t.Errorf("cleanPath should delete at least 1 file, got %d", count)
	}
	if bytes < 1 {
		t.Errorf("cleanPath should free at least 1 byte, got %d", bytes)
	}
}

// Test cleanPath with protected path
func TestCleanPath_ProtectedPath(t *testing.T) {
	// cleanPath doesn't check for protected paths, that's done by the caller
	// This test just verifies it doesn't panic
	tmpDir := t.TempDir()

	count, bytes, err := cleanPath(tmpDir, []string{"*"})
	if err != nil {
		t.Logf("cleanPath returned error (may be expected): %v", err)
	}
	_ = count
	_ = bytes
}

// Test cleanPath with non-existent path
func TestCleanPath_NonExistentPath(t *testing.T) {
	count, bytes, err := cleanPath("/nonexistent/path/that/does/not/exist", []string{"*"})
	if err != nil {
		t.Logf("cleanPath returned error for non-existent path: %v", err)
	}
	if count != 0 {
		t.Errorf("cleanPath should return 0 for non-existent path, got %d", count)
	}
	if bytes != 0 {
		t.Errorf("cleanPath should return 0 bytes for non-existent path, got %d", bytes)
	}
}

// Test cleanPath with specific pattern
func TestCleanPath_SpecificPatternMatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	os.WriteFile(tmpDir+"/test.txt", []byte("test"), 0644)
	os.WriteFile(tmpDir+"/test.log", []byte("log"), 0644)
	os.WriteFile(tmpDir+"/test.dat", []byte("data"), 0644)

	// Clean only .txt files
	count, bytes, err := cleanPath(tmpDir, []string{"*.txt"})
	if err != nil {
		t.Errorf("cleanPath returned error: %v", err)
	}
	if count != 1 {
		t.Errorf("cleanPath should delete 1 file, got %d", count)
	}

	// Verify other files still exist
	if _, err := os.Stat(tmpDir + "/test.log"); os.IsNotExist(err) {
		t.Error("test.log should still exist")
	}
	if _, err := os.Stat(tmpDir + "/test.dat"); os.IsNotExist(err) {
		t.Error("test.dat should still exist")
	}
	_ = bytes
}

// Test CacheCleanerState total size calculation
func TestCacheCleanerState_TotalSizeCalculation(t *testing.T) {
	state := &CacheCleanerState{
		Targets: []*cleaner.CleanTarget{
			{ID: "target-1"},
			{ID: "target-2"},
		},
		SelectedTargets: map[string]bool{
			"target-1": true,
		},
		TotalSize: 100,
	}

	// Simulate size update
	state.TotalSize += 200

	if state.TotalSize != 300 {
		t.Errorf("TotalSize = %d, want 300", state.TotalSize)
	}
}

// Test formatBytes helper (may be in cache_cleaner.go)
func TestCacheCleanerFormatBytes(t *testing.T) {
	// The cache_cleaner.go may have its own formatBytes or use the one from app.go
	// This test verifies the function works correctly
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
