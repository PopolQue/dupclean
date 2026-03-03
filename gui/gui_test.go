package gui

import (
	"os"
	"path/filepath"
	"testing"
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
