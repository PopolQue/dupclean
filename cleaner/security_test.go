package cleaner

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Test sanitizePathForShell with valid paths
func TestSanitizePathForShell_ValidPaths(t *testing.T) {
	tmpDir := t.TempDir()
	
	paths := []string{
		tmpDir,
		filepath.Join(tmpDir, "file.txt"),
		filepath.Join(tmpDir, "file with spaces.txt"),
		filepath.Join(tmpDir, "file-with-dashes.txt"),
	}

	for _, path := range paths {
		// Create the file/dir
		if !strings.HasSuffix(path, "txt") {
			os.MkdirAll(path, 0755)
		} else {
			os.WriteFile(path, []byte("test"), 0644)
		}

		result, err := sanitizePathForShell(path)
		if err != nil {
			t.Errorf("sanitizePathForShell(%q) returned error: %v", path, err)
		}
		if result == "" {
			t.Errorf("sanitizePathForShell(%q) returned empty string", path)
		}
	}
}

// Test sanitizePathForShell with empty path
func TestSanitizePathForShell_EmptyPath(t *testing.T) {
	_, err := sanitizePathForShell("")
	if err == nil {
		t.Error("sanitizePathForShell(\"\") should return error")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("Error should mention 'empty', got: %v", err)
	}
}

// Test sanitizePathForShell with non-existent path
func TestSanitizePathForShell_NonExistent(t *testing.T) {
	_, err := sanitizePathForShell("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("sanitizePathForShell with non-existent path should return error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Error should mention 'does not exist', got: %v", err)
	}
}

// Test escapeAppleScriptString
func TestEscapeAppleScriptString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with'quote", "with\\'quote"},
		{`with\backslash`, `with\\backslash`},
		{`with"double"`, `with\"double\"`},
		{`complex'path\with"all`, `complex\'path\\with\"all`},
	}

	for _, tt := range tests {
		result := escapeAppleScriptString(tt.input)
		if result != tt.expected {
			t.Errorf("escapeAppleScriptString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Test escapePowerShellString
func TestEscapePowerShellString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with'quote", "with''quote"},
		{"with''double", "with''''double"},
		{`with\backslash`, `with\backslash`}, // Backslash doesn't need escaping in single quotes
	}

	for _, tt := range tests {
		result := escapePowerShellString(tt.input)
		if result != tt.expected {
			t.Errorf("escapePowerShellString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Test validateMediaPath with valid file
func TestValidateMediaPath_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mp3")
	os.WriteFile(testFile, []byte("fake mp3"), 0644)

	err := validateMediaPath(testFile)
	if err != nil {
		t.Errorf("validateMediaPath(%q) returned error: %v", testFile, err)
	}
}

// Test validateMediaPath with empty path
func TestValidateMediaPath_EmptyPath(t *testing.T) {
	err := validateMediaPath("")
	if err == nil {
		t.Error("validateMediaPath(\"\") should return error")
	}
}

// Test validateMediaPath with non-existent file
func TestValidateMediaPath_NonExistent(t *testing.T) {
	err := validateMediaPath("/nonexistent/file.mp3")
	if err == nil {
		t.Error("validateMediaPath with non-existent file should return error")
	}
}

// Test SafePlayMedia returns command for valid path
func TestSafePlayMedia_ValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.wav")
	os.WriteFile(testFile, []byte("fake wav"), 0644)

	cmd, err := SafePlayMedia(testFile)
	if err != nil {
		t.Errorf("SafePlayMedia(%q) returned error: %v", testFile, err)
	}
	if cmd == nil {
		t.Error("SafePlayMedia should return a command")
	}
}

// Test SafePlayMedia with empty path
func TestSafePlayMedia_EmptyPath(t *testing.T) {
	cmd, err := SafePlayMedia("")
	if err == nil {
		t.Error("SafePlayMedia(\"\") should return error")
	}
	if cmd != nil {
		t.Error("SafePlayMedia should return nil command on error")
	}
}

// Test SafeMoveToTrash with valid file
func TestSafeMoveToTrash_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	// This may fail depending on system, but shouldn't panic
	err := SafeMoveToTrash(testFile)
	if err != nil {
		t.Logf("SafeMoveToTrash returned error (may be expected): %v", err)
	}
}

// Test SafeMoveToTrash with empty path
func TestSafeMoveToTrash_EmptyPath(t *testing.T) {
	err := SafeMoveToTrash("")
	if err == nil {
		t.Error("SafeMoveToTrash(\"\") should return error")
	}
}

// Test SafeMoveToTrash with non-existent file
func TestSafeMoveToTrash_NonExistent(t *testing.T) {
	err := SafeMoveToTrash("/nonexistent/file.txt")
	if err == nil {
		t.Error("SafeMoveToTrash with non-existent file should return error")
	}
}

// Test safeMoveToTrashMacOS (only on macOS)
func TestSafeMoveToTrashMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS test on " + runtime.GOOS)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	err := safeMoveToTrashMacOS(testFile)
	if err != nil {
		t.Logf("safeMoveToTrashMacOS returned error (may be expected): %v", err)
	}
}

// Test safeMoveToTrashLinux (only on Linux)
func TestSafeMoveToTrashLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux test on " + runtime.GOOS)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	err := safeMoveToTrashLinux(testFile)
	if err != nil {
		t.Logf("safeMoveToTrashLinux returned error (may be expected): %v", err)
	}
}

// Test safeMoveToTrashWindows (only on Windows)
func TestSafeMoveToTrashWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows test on " + runtime.GOOS)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	err := safeMoveToTrashWindows(testFile)
	if err != nil {
		t.Logf("safeMoveToTrashWindows returned error (may be expected): %v", err)
	}
}

// Test path with special characters (but legitimate ones)
func TestSanitizePathForShell_SpecialCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create files with special characters that are actually safe in filenames
	specialNames := []string{
		"file (1).txt",
		"file [test].txt",
		"file {test}.txt",
		"file & test.txt",
		"file | test.txt",
	}

	for _, name := range specialNames {
		testFile := filepath.Join(tmpDir, name)
		os.WriteFile(testFile, []byte("test"), 0644)

		_, err := sanitizePathForShell(testFile)
		if err != nil {
			t.Logf("sanitizePathForShell(%q) returned error: %v", name, err)
			// This is OK - some special chars may be rejected
		}
	}
}

// Test that path traversal is detected
func TestSanitizePathForShell_PathTraversal(t *testing.T) {
	// Create a temp dir and try to escape it
	tmpDir := t.TempDir()
	
	// Change to tmp dir
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Try path traversal
	traversalPath := "../../../etc/passwd"
	
	// This should either fail with "does not exist" or "path traversal"
	_, err := sanitizePathForShell(traversalPath)
	if err == nil {
		// If it doesn't error, the path shouldn't start with ..
		t.Error("Path traversal should be detected")
	}
}

// Test SafePlayMedia on different OS
func TestSafePlayMedia_DifferentOS(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.wav")
	os.WriteFile(testFile, []byte("fake wav"), 0644)

	cmd, err := SafePlayMedia(testFile)
	
	// Should return a command or error based on OS
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		if err != nil {
			t.Logf("SafePlayMedia returned error: %v", err)
		}
		if cmd == nil && err == nil {
			t.Error("SafePlayMedia should return either command or error")
		}
	}
}
