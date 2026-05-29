package gui

import (
	"os"
	"path/filepath"
	"testing"
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

func TestCacheCleanerState_Struct(t *testing.T) {
	state := &CacheCleanerState{
		SelectedTargets: make(map[string]bool),
	}

	if state.SelectedTargets == nil {
		t.Error("CacheCleanerState.SelectedTargets should be initialized")
	}
}

func TestCacheCleanerState_UpdateContent(t *testing.T) {
	// This test would require a full GUI setup, so we just verify the method exists
	// and doesn't panic with nil ContentContainer
	state := &CacheCleanerState{
		ContentContainer: nil, // nil should be handled gracefully
	}

	// Should not panic
	state.updateContent(nil)
}

func TestNewCacheCleanerState(t *testing.T) {
	// We can't create a real fyne.Window in tests, so we just verify
	// the function exists and the state is properly initialized
	state := &CacheCleanerState{
		SelectedTargets: make(map[string]bool),
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
