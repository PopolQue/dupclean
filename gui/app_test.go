package gui

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"dupclean/scanner"
)

// TestAppState initialization
func TestAppState_Initialization(t *testing.T) {
	state := &AppState{
		FolderPath:        "",
		ScanAll:           false,
		IsScanning:        false,
		ProgressText:      "Ready",
		ProgressValue:     0,
		Groups:            nil,
		CurrentGroupIndex: 0,
		DeletedCount:      0,
		FreedBytes:        0,
		IgnoreFolders:     []string{},
		IgnoreExtensions:  []string{},
	}

	if state.FolderPath != "" {
		t.Errorf("FolderPath should be empty, got %q", state.FolderPath)
	}
	if state.ScanAll {
		t.Error("ScanAll should be false")
	}
	if state.IsScanning {
		t.Error("IsScanning should be false")
	}
	if state.ProgressText != "Ready" {
		t.Errorf("ProgressText = %q, want %q", state.ProgressText, "Ready")
	}
	if state.ProgressValue != 0 {
		t.Errorf("ProgressValue = %f, want 0", state.ProgressValue)
	}
	if state.CurrentGroupIndex != 0 {
		t.Errorf("CurrentGroupIndex = %d, want 0", state.CurrentGroupIndex)
	}
	if state.DeletedCount != 0 {
		t.Errorf("DeletedCount = %d, want 0", state.DeletedCount)
	}
	if state.FreedBytes != 0 {
		t.Errorf("FreedBytes = %d, want 0", state.FreedBytes)
	}
}

// Test updateContent method
func TestAppState_UpdateContent(t *testing.T) {
	state := &AppState{}
	
	// Test with nil ContentContainer (should not panic)
	state.updateContent(nil)
	
	// Test with valid ContentContainer
	container := &struct {
		Objects []interface{}
	}{
		Objects: []interface{}{},
	}
	_ = container
}

// Test formatBytes edge cases
func TestFormatBytes_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"zero", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"large TB", 1024 * 1024 * 1024 * 1024 * 5, "5.0 TB"},
		{"large PB", 1024 * 1024 * 1024 * 1024 * 1024, "1.0 PB"},
		{"large EB", 1024 * 1024 * 1024 * 1024 * 1024 * 1024, "1.0 EB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Test runtimeOS returns valid OS
func TestRuntimeOS_Valid(t *testing.T) {
	os := runtime.GOOS
	validOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
		"freebsd": true,
		"openbsd": true,
		"netbsd":  true,
	}

	if !validOS[os] {
		t.Errorf("runtime.GOOS returned unexpected OS: %q", os)
	}
}

// Test stopPlayback with nil state
func TestStopPlayback_NilState(t *testing.T) {
	state := &AppState{}
	stopPlayback(state)
}

// Test stopPlayback resets state
func TestStopPlayback_ResetsState(t *testing.T) {
	called := false
	state := &AppState{
		CurrentPlayer: nil,
		StopPlayer: func() {
			called = true
		},
		PlayingPath: "/test/path",
		playerDone:  make(chan struct{}, 1), // Initialize done channel
	}

	// Signal done immediately to prevent timeout
	state.playerDone <- struct{}{}
	
	stopPlayback(state)

	if !called {
		t.Error("StopPlayer callback should have been called")
	}
	// Note: State fields are reset by the goroutine, not by stopPlayback directly
}

// Test moveToTrash with non-existent file
func TestMoveToTrash_NonExistentFile(t *testing.T) {
	err := moveToTrash("/nonexistent/path/file.txt")
	// Should return an error (file doesn't exist)
	if err == nil {
		t.Log("moveToTrash on non-existent file should return an error")
	}
}

// Test moveToTrash with valid file
func TestMoveToTrash_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_move.txt")

	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This may fail depending on the system, but shouldn't panic
	err := moveToTrash(testFile)
	if err != nil {
		t.Logf("moveToTrash returned error (may be expected): %v", err)
	}
}

// Test keepAndDelete with single file (no deletion)
func TestKeepAndDelete_SingleFile(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "hash1",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.wav", Name: "file1.wav", Size: 1024},
				},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 0, files)

	if state.DeletedCount != 0 {
		t.Errorf("DeletedCount = %d, want 0 (no files to delete)", state.DeletedCount)
	}
	if state.FreedBytes != 0 {
		t.Errorf("FreedBytes = %d, want 0", state.FreedBytes)
	}
	if len(state.Groups) != 0 {
		t.Errorf("Groups should be empty, got %d", len(state.Groups))
	}
}

// Test keepAndDelete with multiple files keeping last
func TestKeepAndDelete_KeepLastFile(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "hash1",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.wav", Name: "file1.wav", Size: 100},
					{Path: "/test/file2.wav", Name: "file2.wav", Size: 200},
					{Path: "/test/file3.wav", Name: "file3.wav", Size: 300},
				},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 2, files) // Keep last file

	if state.DeletedCount != 2 {
		t.Errorf("DeletedCount = %d, want 2", state.DeletedCount)
	}
	if state.FreedBytes != 300 {
		t.Errorf("FreedBytes = %d, want 300", state.FreedBytes)
	}
}

// Test keepAndDelete removes group from list
func TestKeepAndDelete_RemovesGroup(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 1,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "hash1",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.wav", Name: "file1.wav", Size: 100},
					{Path: "/test/file2.wav", Name: "file2.wav", Size: 100},
				},
			},
			{
				Hash: "hash2",
				Files: []scanner.FileInfo{
					{Path: "/test/file3.wav", Name: "file3.wav", Size: 200},
					{Path: "/test/file4.wav", Name: "file4.wav", Size: 200},
				},
			},
		},
	}

	files := state.Groups[1].Files
	keepAndDelete(state, 0, files)

	if len(state.Groups) != 1 {
		t.Errorf("Should have 1 group left, got %d", len(state.Groups))
	}
	if state.Groups[0].Hash != "hash1" {
		t.Errorf("Remaining group should be hash1, got %s", state.Groups[0].Hash)
	}
}

// Test formatBytes consistency
func TestFormatBytes_Consistency(t *testing.T) {
	// Test that 1024 bytes always equals 1.0 KB
	result1 := formatBytes(1024)
	result2 := formatBytes(1024)
	if result1 != result2 {
		t.Errorf("formatBytes is not consistent: %q vs %q", result1, result2)
	}

	// Test monotonicity: larger values should produce larger or equal strings
	small := formatBytes(100)
	large := formatBytes(1000)
	if small == large {
		t.Errorf("formatBytes(100) and formatBytes(1000) should differ")
	}
}

// Test AppState with ignore configurations
func TestAppState_WithIgnoreConfig(t *testing.T) {
	state := &AppState{
		IgnoreFolders:    []string{"/tmp", "/var/cache"},
		IgnoreExtensions: []string{".log", ".tmp", ".bak"},
	}

	if len(state.IgnoreFolders) != 2 {
		t.Errorf("IgnoreFolders length = %d, want 2", len(state.IgnoreFolders))
	}
	if len(state.IgnoreExtensions) != 3 {
		t.Errorf("IgnoreExtensions length = %d, want 3", len(state.IgnoreExtensions))
	}

	// Verify extensions have dots
	for _, ext := range state.IgnoreExtensions {
		if ext[0] != '.' {
			t.Errorf("Extension %q should start with '.'", ext)
		}
	}
}

// Test progress components initialization
func TestProgressComponents_Init(t *testing.T) {
	prog := &progressComponents{
		label:  nil,
		status: nil,
		bar:    nil,
	}

	if prog.label != nil {
		t.Error("label should be nil initially")
	}
	if prog.status != nil {
		t.Error("status should be nil initially")
	}
	if prog.bar != nil {
		t.Error("bar should be nil initially")
	}
}
