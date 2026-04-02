package gui

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"dupclean/scanner"
)

func TestCleanPath_NonExistent_BusinessLogic(t *testing.T) {
	_, _, err := cleanPath("/nonexistent/path/that/does/not/exist", []string{"*"})

	if err != nil {
		t.Errorf("Expected no error for non-existent path, got %v", err)
	}
}

func TestCleanPath_StarPattern_BusinessLogic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	writeTestFile(t, tmpDir, "file1.txt", "content1")
	writeTestFile(t, tmpDir, "file2.txt", "content2")
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)
	writeTestFile(t, subDir, "file3.txt", "content3")

	count, _, err := cleanPath(tmpDir, []string{"*"})

	if err != nil {
		t.Errorf("cleanPath() error = %v", err)
	}
	if count == 0 {
		t.Error("Expected files to be deleted")
	}

	// Verify base directory still exists
	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Base directory should still exist")
	}
}

func TestCleanPath_SpecificPattern_BusinessLogic(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files with different extensions
	writeTestFile(t, tmpDir, "cache1.tmp", "temp content")
	writeTestFile(t, tmpDir, "cache2.tmp", "temp content")
	writeTestFile(t, tmpDir, "important.txt", "important content")

	count, _, err := cleanPath(tmpDir, []string{"*.tmp"})

	if err != nil {
		t.Errorf("cleanPath() error = %v", err)
	}
	if count == 0 {
		t.Error("Expected .tmp files to be deleted")
	}

	// Verify important.txt still exists
	if _, err := os.Stat(filepath.Join(tmpDir, "important.txt")); os.IsNotExist(err) {
		t.Error("important.txt should still exist")
	}
}

func TestCleanPath_EmptyPatterns_BusinessLogic(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, tmpDir, "file.txt", "content")

	count, _, err := cleanPath(tmpDir, []string{})

	if err != nil {
		t.Errorf("Expected no error for empty patterns, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 files deleted, got %d", count)
	}
}

func TestCleanPath_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	writeTestFile(t, tmpDir, "file.txt", "content")

	count, _, err := cleanPath(tmpDir, []string{"*.xyz"})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 files deleted, got %d", count)
	}

	// Verify file still exists
	if _, err := os.Stat(filepath.Join(tmpDir, "file.txt")); os.IsNotExist(err) {
		t.Error("file.txt should still exist")
	}
}

func TestFormatBytes_BusinessLogic(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestRuntimeOS_BusinessLogic(t *testing.T) {
	os := runtime.GOOS
	if os == "" {
		t.Error("runtime.GOOS should not return empty string")
	}
}

func TestStopPlayback_NoPlayer(t *testing.T) {
	state := &AppState{
		Groups: []scanner.DuplicateGroup{},
	}

	// Should not panic
	stopPlayback(state)
}

func TestKeepAndDelete_SingleGroup(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	writeTestFile(t, tmpDir, "file1.txt", "content1")
	writeTestFile(t, tmpDir, "file2.txt", "content2")

	state := &AppState{
		Groups: []scanner.DuplicateGroup{
			{
				Files: []scanner.FileInfo{
					{Path: file1},
					{Path: file2},
				},
			},
		},
		CurrentGroupIndex: 0,
	}

	// Keep first file, delete second
	keepAndDelete(state, 0, state.Groups[0].Files)

	// Verify first file still exists
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Error("Kept file should still exist")
	}

	// Verify group is removed
	if len(state.Groups) != 0 {
		t.Errorf("Expected 0 groups after deletion, got %d", len(state.Groups))
	}
}

func TestKeepAndDelete_MultipleGroups(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	file3 := filepath.Join(tmpDir, "file3.txt")
	file4 := filepath.Join(tmpDir, "file4.txt")

	writeTestFile(t, tmpDir, "file1.txt", "content1")
	writeTestFile(t, tmpDir, "file2.txt", "content2")
	writeTestFile(t, tmpDir, "file3.txt", "content3")
	writeTestFile(t, tmpDir, "file4.txt", "content4")

	state := &AppState{
		Groups: []scanner.DuplicateGroup{
			{
				Files: []scanner.FileInfo{
					{Path: file1},
					{Path: file2},
				},
			},
			{
				Files: []scanner.FileInfo{
					{Path: file3},
					{Path: file4},
				},
			},
		},
		CurrentGroupIndex: 0,
	}

	// Keep first file in first group
	keepAndDelete(state, 0, state.Groups[0].Files)

	// Should still have second group
	if len(state.Groups) != 1 {
		t.Errorf("Expected 1 group remaining, got %d", len(state.Groups))
	}

	// CurrentGroupIndex should be updated
	if state.CurrentGroupIndex != 0 {
		t.Errorf("Expected CurrentGroupIndex to be 0, got %d", state.CurrentGroupIndex)
	}
}

// Note: showIgnoreDialog requires a GUI window and cannot be tested headlessly

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}
