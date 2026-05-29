package gui

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"dupclean/internal/fsutil"
	"dupclean/scanner"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
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
			result := fsutil.FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("fsutil.FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
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
				Hash: "abc",
				Files: []scanner.FileInfo{
					{Path: file1, Hash: "abc"},
					{Path: file2, Hash: "abc"},
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
				Hash: "abc",
				Files: []scanner.FileInfo{
					{Path: file1, Hash: "abc"},
					{Path: file2, Hash: "abc"},
				},
			},
			{
				Hash: "def",
				Files: []scanner.FileInfo{
					{Path: file3, Hash: "def"},
					{Path: file4, Hash: "def"},
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

func TestSmartCleanAll_BusinessLogic(t *testing.T) {
	oldTrash := moveToTrash
	oldSafe := safeToDelete
	moveToTrash = func(path string) error { return nil }
	safeToDelete = func(f scanner.FileInfo) (bool, error) { return true, nil }
	defer func() {
		moveToTrash = oldTrash
		safeToDelete = oldSafe
	}()

	state := &AppState{
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "abc",
				Files: []scanner.FileInfo{
					{Path: "/path/1/a", Size: 100, Hash: "abc"},
					{Path: "/path/1/b", Size: 100, Hash: "abc"},
				},
			},
			{
				Hash: "def",
				Files: []scanner.FileInfo{
					{Path: "/path/2/a", Size: 200, Hash: "def"},
					{Path: "/path/2/b", Size: 200, Hash: "def"},
				},
			},
		},
	}

	SmartCleanAll(state)

	if state.DeletedCount != 2 {
		t.Errorf("Expected 2 files deleted, got %d", state.DeletedCount)
	}
	if state.FreedBytes != 300 {
		t.Errorf("Expected 300 bytes freed, got %d", state.FreedBytes)
	}
	if len(state.Groups) != 0 {
		t.Errorf("Expected 0 groups left, got %d", len(state.Groups))
	}
}

func TestCleanSelected_BusinessLogic(t *testing.T) {
	oldTrash := moveToTrash
	oldSafe := safeToDelete
	moveToTrash = func(path string) error { return nil }
	safeToDelete = func(f scanner.FileInfo) (bool, error) { return true, nil }
	defer func() {
		moveToTrash = oldTrash
		safeToDelete = oldSafe
	}()

	state := &AppState{
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "abc",
				Files: []scanner.FileInfo{
					{Name: "a", Path: "/path/1/a", Size: 100, Hash: "abc"},
					{Name: "b", Path: "/path/1/b", Size: 100, Hash: "abc"},
				},
			},
		},
		Selections: [][]bool{
			{true, false}, // Keep first, delete second
		},
		ContentContainer: container.NewMax(),
	}

	cleanSelected(state)

	if state.DeletedCount != 1 {
		t.Errorf("Expected 1 file deleted, got %d", state.DeletedCount)
	}
	if state.FreedBytes != 100 {
		t.Errorf("Expected 100 bytes freed, got %d", state.FreedBytes)
	}
}

func TestKeepAndDeleteLocked_Failures(t *testing.T) {
	oldTrash := moveToTrash
	oldSafe := safeToDelete
	defer func() {
		moveToTrash = oldTrash
		safeToDelete = oldSafe
	}()

	state := &AppState{}
	files := []scanner.FileInfo{
		{Name: "a", Path: "/a", Size: 100, ModTime: time.Now(), Hash: "abc"},
		{Name: "b", Path: "/b", Size: 100, ModTime: time.Now(), Hash: "abc"},
	}

	t.Run("SafeToDeleteFail", func(t *testing.T) {
		safeToDelete = func(f scanner.FileInfo) (bool, error) { return false, errors.New("modified") }
		state.SkippedCount = 0
		state.SkippedFiles = nil
		keepAndDeleteLocked(state, 0, files)
		if state.SkippedCount != 1 {
			t.Errorf("Expected 1 skipped, got %d", state.SkippedCount)
		}
	})

	t.Run("MoveToTrashFail", func(t *testing.T) {
		safeToDelete = func(f scanner.FileInfo) (bool, error) { return true, nil }
		moveToTrash = func(path string) error { return errors.New("trash error") }
		state.SkippedCount = 0
		state.SkippedFiles = nil
		state.DeletedCount = 0
		keepAndDeleteLocked(state, 0, files)
		if state.SkippedCount != 1 {
			t.Errorf("Expected 1 skipped, got %d", state.SkippedCount)
		}
		if state.DeletedCount != 0 {
			t.Errorf("Expected 0 deleted, got %d", state.DeletedCount)
		}
	})
}

func TestStartScan_BusinessLogic(t *testing.T) {
	oldFind := findDuplicates
	findDuplicates = func(root string, scanAll bool, progress func(scanner.ScanProgress), ignoreFolders, ignoreExtensions []string) ([]scanner.DuplicateGroup, scanner.ScanStats, error) {
		return []scanner.DuplicateGroup{
			{Hash: "abc", Files: []scanner.FileInfo{{Path: "/a", Hash: "abc"}, {Path: "/b", Hash: "abc"}}},
		}, scanner.ScanStats{TotalScanned: 10}, nil
	}
	defer func() { findDuplicates = oldFind }()

	state := &AppState{
		FolderPath: t.TempDir(),
		progressComponents: &progressComponents{
			label:  widget.NewLabel(""),
			status: widget.NewLabel(""),
			bar:    widget.NewProgressBar(),
		},
		ContentContainer: container.NewMax(),
	}

	startScan(state, nil, nil)

	// Wait for goroutine to finish (heuristic)
	time.Sleep(100 * time.Millisecond)

	state.mu.RLock()
	defer state.mu.RUnlock()
	if len(state.Groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(state.Groups))
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
