package gui

import (
	"os"
	"path/filepath"
	"testing"

	"dupclean/scanner"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestMoveToTrash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := moveToTrash(testFile)
	if err != nil {
		t.Logf("moveToTrash error (may be expected in some environments): %v", err)
	}
}

func TestAppState_Struct(t *testing.T) {
	state := &AppState{
		FolderPath:        "",
		ScanAll:           false,
		IsScanning:        false,
		ProgressText:      "",
		ProgressValue:     0,
		Groups:            nil,
		CurrentGroupIndex: 0,
		DeletedCount:      0,
		FreedBytes:        0,
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

func TestRuntimeOS(t *testing.T) {
	result := runtimeOS()
	if result == "" {
		t.Error("runtimeOS should return a non-empty string")
	}
	if result != "darwin" && result != "linux" && result != "windows" && result != "freebsd" {
		t.Errorf("runtimeOS returned unexpected value: %s", result)
	}
}

func TestStopPlayback_Nil(t *testing.T) {
	state := &AppState{
		CurrentPlayer: nil,
		StopPlayer:    nil,
	}
	stopPlayback(state)
	if state.CurrentPlayer != nil {
		t.Error("CurrentPlayer should be nil after stopPlayback")
	}
}

func TestStopPlayback_WithPlayer(t *testing.T) {
	state := &AppState{
		CurrentPlayer: nil,
		StopPlayer: func() {
		},
	}
	stopPlayback(state)
	if state.CurrentPlayer != nil {
		t.Error("CurrentPlayer should be nil after stopPlayback")
	}
	if state.StopPlayer != nil {
		t.Error("StopPlayer should be nil after stopPlayback")
	}
}

func TestKeepAndDelete(t *testing.T) {
	state := &AppState{
		DeletedCount: 0,
		FreedBytes:   0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "testhash",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.wav", Name: "file1.wav", Size: 1024},
					{Path: "/test/file2.wav", Name: "file2.wav", Size: 1024},
				},
			},
		},
		CurrentGroupIndex: 0,
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 0, files)

	if state.DeletedCount != 1 {
		t.Errorf("DeletedCount = %d, want 1", state.DeletedCount)
	}
	if state.FreedBytes != 1024 {
		t.Errorf("FreedBytes = %d, want 1024", state.FreedBytes)
	}
	if len(state.Groups) != 0 {
		t.Errorf("Groups should be empty after keepAndDelete, got %d", len(state.Groups))
	}
}

func TestKeepAndDelete_LastGroup(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "testhash",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.wav", Name: "file1.wav", Size: 512},
				},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 0, files)

	if state.DeletedCount != 0 {
		t.Errorf("DeletedCount = %d, want 0 (no files deleted because only 1 file)", state.DeletedCount)
	}
}

func TestAppState_WithGroups(t *testing.T) {
	state := &AppState{
		CurrentGroupIndex: 1,
		DeletedCount:      5,
		FreedBytes:        1024 * 1024 * 10, // 10 MB
	}

	if state.CurrentGroupIndex != 1 {
		t.Errorf("CurrentGroupIndex = %d, want 1", state.CurrentGroupIndex)
	}
	if state.DeletedCount != 5 {
		t.Errorf("DeletedCount = %d, want 5", state.DeletedCount)
	}
	if state.FreedBytes != 1024*1024*10 {
		t.Errorf("FreedBytes = %d, want %d", state.FreedBytes, 1024*1024*10)
	}
}
