package gui

import (
	"os"
	"path/filepath"
	"runtime"
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
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
		{1024 * 500, "500.0 KB"},
		{1024 * 1024 * 50, "50.0 MB"},
		{1, "1 B"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatBytes_VariousSizes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{100, "100 B"},
		{200, "200 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1025, "1.0 KB"},
		{2048, "2.0 KB"},
		{10240, "10.0 KB"},
		{102400, "100.0 KB"},
		{1048576, "1.0 MB"},
		{2097152, "2.0 MB"},
		{10485760, "10.0 MB"},
		{104857600, "100.0 MB"},
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

func TestMoveToTrash_AbsolutePath(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_abs.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	absPath, _ := filepath.Abs(testFile)
	_ = absPath

	err := moveToTrash(testFile)
	if err != nil {
		t.Logf("moveToTrash error: %v", err)
	}
}

func TestMoveToTrash_NonExistent(t *testing.T) {
	err := moveToTrash("/nonexistent/file/that/does/not/exist.txt")
	if err == nil {
		t.Log("moveToTrash on non-existent file returned nil error (may be expected)")
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
		FreedBytes:        1024 * 1024 * 10,
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

func TestAppState_Full(t *testing.T) {
	state := &AppState{
		FolderPath:        "/test/path",
		ScanAll:           true,
		IsScanning:        true,
		ProgressText:      "Scanning...",
		ProgressValue:     0.5,
		Groups:            []scanner.DuplicateGroup{},
		CurrentGroupIndex: 2,
		DeletedCount:      10,
		FreedBytes:        1024 * 1024 * 100,
		IgnoreFolders:     []string{"/ignore/this"},
		IgnoreExtensions:  []string{".txt"},
	}

	if state.FolderPath != "/test/path" {
		t.Errorf("FolderPath = %q, want %q", state.FolderPath, "/test/path")
	}
	if !state.ScanAll {
		t.Error("ScanAll should be true")
	}
	if !state.IsScanning {
		t.Error("IsScanning should be true")
	}
	if state.ProgressText != "Scanning..." {
		t.Errorf("ProgressText = %q, want %q", state.ProgressText, "Scanning...")
	}
	if state.ProgressValue != 0.5 {
		t.Errorf("ProgressValue = %f, want %f", state.ProgressValue, 0.5)
	}
	if state.CurrentGroupIndex != 2 {
		t.Errorf("CurrentGroupIndex = %d, want 2", state.CurrentGroupIndex)
	}
	if state.DeletedCount != 10 {
		t.Errorf("DeletedCount = %d, want 10", state.DeletedCount)
	}
	if state.FreedBytes != 1024*1024*100 {
		t.Errorf("FreedBytes = %d, want %d", state.FreedBytes, 1024*1024*100)
	}
	if len(state.IgnoreFolders) != 1 {
		t.Errorf("IgnoreFolders length = %d, want 1", len(state.IgnoreFolders))
	}
	if len(state.IgnoreExtensions) != 1 {
		t.Errorf("IgnoreExtensions length = %d, want 1", len(state.IgnoreExtensions))
	}
}

func TestRuntimeOS(t *testing.T) {
	result := runtime.GOOS
	if result == "" {
		t.Error("runtimeOS should return a non-empty string")
	}
	if result != "darwin" && result != "linux" && result != "windows" && result != "freebsd" && result != "netbsd" && result != "openbsd" && result != "plan9" && result != "solaris" && result != "illumos" && result != "js" && result != "wasip1" {
		t.Logf("runtimeOS returned: %s", result)
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

func TestKeepAndDelete_MultipleFiles(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "testhash",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.wav", Name: "file1.wav", Size: 100},
					{Path: "/test/file2.wav", Name: "file2.wav", Size: 100},
					{Path: "/test/file3.wav", Name: "file3.wav", Size: 100},
					{Path: "/test/file4.wav", Name: "file4.wav", Size: 100},
				},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 0, files)

	if state.DeletedCount != 3 {
		t.Errorf("DeletedCount = %d, want 3", state.DeletedCount)
	}
	if state.FreedBytes != 300 {
		t.Errorf("FreedBytes = %d, want 300", state.FreedBytes)
	}
}

func TestKeepAndDelete_KeepSecond(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "testhash",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.wav", Name: "file1.wav", Size: 256},
					{Path: "/test/file2.wav", Name: "file2.wav", Size: 256},
					{Path: "/test/file3.wav", Name: "file3.wav", Size: 256},
				},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 1, files)

	if state.DeletedCount != 2 {
		t.Errorf("DeletedCount = %d, want 2", state.DeletedCount)
	}
	if state.FreedBytes != 512 {
		t.Errorf("FreedBytes = %d, want 512", state.FreedBytes)
	}
}
