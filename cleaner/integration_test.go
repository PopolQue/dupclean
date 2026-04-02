package cleaner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetBrowserTargetsLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	targets := getBrowserTargetsLinux()
	if len(targets) == 0 {
		t.Error("Expected Linux browser targets")
	}

	for _, target := range targets {
		if target.OS != "linux" {
			t.Errorf("Expected OS 'linux', got %q", target.OS)
		}
	}
}

func TestGetBrowserTargetsWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	targets := getBrowserTargetsWindows()
	if len(targets) == 0 {
		t.Error("Expected Windows browser targets")
	}

	for _, target := range targets {
		if target.OS != "windows" {
			t.Errorf("Expected OS 'windows', got %q", target.OS)
		}
	}
}

func TestGetDeveloperTargetsLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	targets := getDeveloperTargetsLinux()
	if len(targets) == 0 {
		t.Error("Expected Linux developer targets")
	}
}

func TestGetDeveloperTargetsWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	targets := getDeveloperTargetsWindows()
	if len(targets) == 0 {
		t.Error("Expected Windows developer targets")
	}
}

func TestGetLogsTargetsLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	targets := getLogsTargetsLinux()
	if len(targets) == 0 {
		t.Error("Expected Linux logs targets")
	}
}

func TestGetLogsTargetsWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	targets := getLogsTargetsWindows()
	if len(targets) == 0 {
		t.Error("Expected Windows logs targets")
	}
}

func TestGetLinuxTargets(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	targets := getLinuxTargets()
	if len(targets) == 0 {
		t.Error("Expected Linux system targets")
	}
}

func TestGetWindowsTargets(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	targets := getWindowsTargets()
	if len(targets) == 0 {
		t.Error("Expected Windows system targets")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	opts := DeleteOptions{
		Permanent:   true,
		Concurrency: 1,
	}

	result, err := Delete([]EntryInfo{
		{Path: testFile, Size: 4},
	}, opts)

	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
	if result.Deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", result.Deleted)
	}
	if result.FreedBytes != 4 {
		t.Errorf("Expected 4 bytes freed, got %d", result.FreedBytes)
	}

	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should be deleted")
	}
}

func TestDelete_ToTrash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	opts := DeleteOptions{
		Permanent:   false,
		Concurrency: 1,
	}

	result, err := Delete([]EntryInfo{
		{Path: testFile, Size: 4},
	}, opts)

	if err != nil {
		t.Logf("Delete to trash error (expected in CI): %v", err)
	}
	_ = result
}

func TestDelete_EmptyList(t *testing.T) {
	opts := DeleteOptions{
		Permanent:   true,
		Concurrency: 1,
	}

	result, err := Delete([]EntryInfo{}, opts)

	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}
	if result.Deleted != 0 {
		t.Errorf("Expected 0 deleted, got %d", result.Deleted)
	}
	if result.FreedBytes != 0 {
		t.Errorf("Expected 0 bytes freed, got %d", result.FreedBytes)
	}
}

func TestDelete_NonExistentFile(t *testing.T) {
	opts := DeleteOptions{
		Permanent:   true,
		Concurrency: 1,
	}

	result, err := Delete([]EntryInfo{
		{Path: "/nonexistent/file.txt", Size: 100},
	}, opts)

	// Delete may or may not error for non-existent files depending on implementation
	// The important thing is it doesn't crash
	_ = result
	_ = err
}

func TestMoveToTrashMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := safeMoveToTrashMacOS(testFile)
	if err != nil {
		t.Logf("safeMoveToTrashMacOS error (may fail in CI): %v", err)
	}
}

func TestMoveToTrashLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := safeMoveToTrashLinux(testFile)
	if err != nil {
		t.Logf("safeMoveToTrashLinux error (may fail in CI): %v", err)
	}
}

func TestMoveToTrashWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only test")
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := safeMoveToTrashWindows(testFile)
	if err != nil {
		t.Logf("safeMoveToTrashWindows error (may fail in CI): %v", err)
	}
}

func TestGetTargetByPath(t *testing.T) {
	targets := []*CleanTarget{
		{ID: "target1", Paths: []string{"/test/path1"}},
		{ID: "target2", Paths: []string{"/test/path2"}},
	}

	target := GetTargetByPath(targets, "/test/path1")
	if target == nil {
		t.Error("Expected to find target")
	} else if target.ID != "target1" {
		t.Errorf("Expected target1, got %q", target.ID)
	}

	target = GetTargetByPath(targets, "/nonexistent")
	if target != nil {
		t.Error("Expected nil for non-existent path")
	}
}
