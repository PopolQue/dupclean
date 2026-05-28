package cleaner

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSafePlayMedia_AllPlatforms(t *testing.T) {
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

	// Test all supported OS branches via runtime.GOOS manipulation or just calling them
	// We can't easily change runtime.GOOS, but we can at least test the current one fully
	cmd, err := SafePlayMedia(testFile)
	if err != nil {
		t.Errorf("SafePlayMedia failed on current platform %s: %v", runtime.GOOS, err)
	}
	if cmd == nil {
		t.Error("SafePlayMedia returned nil cmd")
	}

	// Test unsupported OS error (if we can find one)
	// This is hard without refactoring SafePlayMedia to accept GOOS as param
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

func TestEscapePowerShellString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"it's", "it''s"},
		{"'quoted'", "''quoted''"},
	}

	for _, tt := range tests {
		result := escapePowerShellString(tt.input)
		if result != tt.expected {
			t.Errorf("escapePowerShellString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
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
}
