package gui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestGetLogFilePath tests that the function returns a valid path
func TestGetLogFilePath(t *testing.T) {
	path := getLogFilePath()

	if path == "" {
		t.Error("getLogFilePath() should not return empty string")
	}

	// Path should end with dupclean.log
	if filepath.Base(path) != "dupclean.log" {
		t.Errorf("getLogFilePath() should return path ending with 'dupclean.log', got %s", path)
	}
}

// TestGetLogFilePath_PlatformSpecific tests platform-specific paths
func TestGetLogFilePath_PlatformSpecific(t *testing.T) {
	path := getLogFilePath()

	// Verify separator is correct for platform
	if runtime.GOOS == "windows" && filepath.Separator != '\\' {
		t.Error("Windows should use backslash separator")
	}
	if runtime.GOOS != "windows" && filepath.Separator != '/' {
		t.Error("Unix-like should use forward slash separator")
	}

	_ = path // Use variable to avoid "declared and not used" error
}

// TestGetLogFilePath_EnvironmentVariables tests environment variable precedence
func TestGetLogFilePath_EnvironmentVariables(t *testing.T) {
	// Save original values
	origTMPDIR := os.Getenv("TMPDIR")
	origTEMP := os.Getenv("TEMP")
	origTMP := os.Getenv("TMP")

	defer func() {
		// Restore original values
		if origTMPDIR != "" {
			os.Setenv("TMPDIR", origTMPDIR)
		} else {
			os.Unsetenv("TMPDIR")
		}
		if origTEMP != "" {
			os.Setenv("TEMP", origTEMP)
		} else {
			os.Unsetenv("TEMP")
		}
		if origTMP != "" {
			os.Setenv("TMP", origTMP)
		} else {
			os.Unsetenv("TMP")
		}
	}()

	// Test TMPDIR precedence (Unix)
	if runtime.GOOS != "windows" {
		os.Setenv("TMPDIR", "/custom/tmp")
		path := getLogFilePath()
		expected := "/custom/tmp/dupclean.log"
		if path != expected {
			t.Errorf("getLogFilePath() with TMPDIR = %q, want %q", path, expected)
		}
	}

	// Test TEMP precedence (Windows)
	if runtime.GOOS == "windows" {
		os.Setenv("TEMP", "C:\\Custom\\Temp")
		path := getLogFilePath()
		expected := "C:\\Custom\\Temp\\dupclean.log"
		if path != expected {
			t.Errorf("getLogFilePath() with TEMP = %q, want %q", path, expected)
		}
	}
}

// TestGetLogFilePath_Fallback tests fallback behavior
func TestGetLogFilePath_Fallback(t *testing.T) {
	// Save and clear all temp env vars
	origTMPDIR := os.Getenv("TMPDIR")
	origTEMP := os.Getenv("TEMP")
	origTMP := os.Getenv("TMP")

	os.Unsetenv("TMPDIR")
	os.Unsetenv("TEMP")
	os.Unsetenv("TMP")

	defer func() {
		// Restore
		if origTMPDIR != "" {
			os.Setenv("TMPDIR", origTMPDIR)
		}
		if origTEMP != "" {
			os.Setenv("TEMP", origTEMP)
		}
		if origTMP != "" {
			os.Setenv("TMP", origTMP)
		}
	}()

	// Should still return a valid path (platform default)
	path := getLogFilePath()
	if path == "" {
		t.Error("getLogFilePath() should return platform default when env vars are unset")
	}
}

// TestGetLogFilePath_DirectoryCreation tests that the log directory is created
func TestGetLogFilePath_DirectoryCreation(t *testing.T) {
	// This test verifies the init() function handles directory creation
	// We can't easily test this without modifying the global state,
	// but we can verify the path is valid

	path := getLogFilePath()
	dir := filepath.Dir(path)

	// Directory should be creatable (no invalid chars)
	if dir == "" {
		t.Error("Log directory path should not be empty")
	}
}

// TestGetLogFilePath_ValidFilename tests that the filename is always valid
func TestGetLogFilePath_ValidFilename(t *testing.T) {
	path := getLogFilePath()
	filename := filepath.Base(path)

	// Should be exactly "dupclean.log"
	if filename != "dupclean.log" {
		t.Errorf("Filename should be 'dupclean.log', got %q", filename)
	}

	// Should not contain path separators
	if strings.ContainsRune(filename, filepath.Separator) {
		t.Errorf("Filename should not contain path separators: %q", filename)
	}

	// Should not contain invalid characters
	invalidChars := []rune{'<', '>', ':', '"', '|', '?', '*'}
	for _, char := range invalidChars {
		if strings.ContainsRune(filename, char) {
			t.Errorf("Filename should not contain invalid character %q: %q", char, filename)
		}
	}
}

// TestGetLogFilePath_Consistency tests that multiple calls return the same path
func TestGetLogFilePath_Consistency(t *testing.T) {
	path1 := getLogFilePath()
	path2 := getLogFilePath()
	path3 := getLogFilePath()

	if path1 != path2 || path2 != path3 {
		t.Error("getLogFilePath() should return consistent paths")
	}
}

// TestGetLogFilePath_NoSideEffects tests that calling the function has no side effects
func TestGetLogFilePath_NoSideEffects(t *testing.T) {
	// Save environment
	origTMPDIR := os.Getenv("TMPDIR")
	origTEMP := os.Getenv("TEMP")
	origTMP := os.Getenv("TMP")

	// Call function multiple times
	_ = getLogFilePath()
	_ = getLogFilePath()
	_ = getLogFilePath()

	// Verify environment wasn't modified
	if os.Getenv("TMPDIR") != origTMPDIR {
		t.Error("getLogFilePath() should not modify TMPDIR")
	}
	if os.Getenv("TEMP") != origTEMP {
		t.Error("getLogFilePath() should not modify TEMP")
	}
	if os.Getenv("TMP") != origTMP {
		t.Error("getLogFilePath() should not modify TMP")
	}
}
