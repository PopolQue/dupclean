package cleaner

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSafePlayMedia_Logic(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.wav")
	if err := os.WriteFile(testFile, []byte("dummy wave content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock execCommand to avoid actually running anything
	oldExec := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() { execCommand = oldExec }()

	oldOS := goos
	defer func() { goos = oldOS }()

	platforms := []string{"darwin", "linux", "windows", "unsupported"}

	for _, p := range platforms {
		t.Run("Platform_"+p, func(t *testing.T) {
			goos = p
			cmd, err := SafePlayMedia(testFile)

			if p == "unsupported" {
				if err == nil {
					t.Errorf("Expected error for unsupported platform")
				}
				return
			}

			if err != nil {
				t.Errorf("SafePlayMedia failed on platform %s: %v", p, err)
			}
			if cmd == nil {
				t.Error("SafePlayMedia returned nil cmd")
			}
		})
	}
}

func TestSafeMoveToTrash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_trash.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wrapper for internal/trash, just ensure it doesn't crash
	_ = SafeMoveToTrash(testFile)
}

func TestValidateMediaPath(t *testing.T) {

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.wav")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := validateMediaPath(testFile); err != nil {
		t.Errorf("validateMediaPath failed for valid file: %v", err)
	}

	if err := validateMediaPath(""); err == nil {
		t.Error("validateMediaPath should fail for empty path")
	}

	if err := validateMediaPath(filepath.Join(tmpDir, "missing.wav")); err == nil {
		t.Error("validateMediaPath should fail for missing file")
	}

	// Test absPath error
	oldAbs := absPath
	absPath = func(path string) (string, error) { return "", errors.New("abs error") }
	defer func() { absPath = oldAbs }()

	if err := validateMediaPath("foo"); err == nil {
		t.Error("validateMediaPath should fail on abs error")
	}
}
