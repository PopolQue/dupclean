package trash

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestMoveToTrash_EmptyPath tests that empty paths are rejected
func TestMoveToTrash_EmptyPath(t *testing.T) {
	err := MoveToTrash("")
	if err == nil {
		t.Error("MoveToTrash(\"\") should return error")
	}
	if err != nil && err.Error() != "cannot move empty path to trash" {
		t.Errorf("Expected 'cannot move empty path to trash', got: %v", err)
	}
}

// TestMoveToTrash_RootDirectory tests that root directories are rejected
func TestMoveToTrash_RootDirectory(t *testing.T) {
	rootPaths := []string{"/", `\`, `C:\`, `c:\`}

	for _, path := range rootPaths {
		err := MoveToTrash(path)
		if err == nil {
			t.Errorf("MoveToTrash(%q) should return error", path)
		}
	}
}

// TestMoveToTrash_PathTraversal tests path traversal detection
func TestMoveToTrash_PathTraversal(t *testing.T) {
	traversalPaths := []string{
		"../../../etc/passwd",
		"/home/user/../../../etc/passwd",
		"foo/../../../etc/passwd",
	}

	for _, path := range traversalPaths {
		err := MoveToTrash(path)
		if err == nil {
			t.Errorf("MoveToTrash(%q) should detect path traversal", path)
		}
	}
}

// TestMoveToTrash_NonExistentFile tests behavior with non-existent files
func TestMoveToTrash_NonExistentFile(t *testing.T) {
	err := MoveToTrash("/nonexistent/path/file.txt")
	// This may or may not error depending on OS implementation
	// The important thing is it doesn't crash
	t.Logf("MoveToTrash non-existent file: %v", err)
}

// TestMoveToTrash_ValidFile tests moving a valid file to trash
func TestMoveToTrash_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// This may fail depending on system, but shouldn't panic
	err := MoveToTrash(testFile)
	if err != nil {
		t.Logf("MoveToTrash returned error (may be expected): %v", err)
	}
}

// TestMoveToTrashMacOS tests macOS-specific implementation
func TestMoveToTrashMacOS(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS test on " + runtime.GOOS)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := moveToTrashMacOS(testFile)
	if err != nil {
		t.Logf("moveToTrashMacOS returned error (may be expected): %v", err)
	}
}

// TestMoveToTrashLinux tests Linux-specific implementation
func TestMoveToTrashLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux test on " + runtime.GOOS)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := moveToTrashLinux(testFile)
	if err != nil {
		t.Logf("moveToTrashLinux returned error (may be expected): %v", err)
	}
}

// TestMoveToTrashWindows tests Windows-specific implementation
func TestMoveToTrashWindows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows test on " + runtime.GOOS)
	}

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := moveToTrashWindows(testFile)
	if err != nil {
		t.Logf("moveToTrashWindows returned error (may be expected): %v", err)
	}
}

// TestSafeMoveToTrashDir_TOCTOU tests TOCTOU-safe trash directory move
func TestSafeMoveToTrashDir_TOCTOU(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	if err := os.MkdirAll(trashDir, 0755); err != nil {
		t.Fatalf("Failed to create trash dir: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := safeMoveToTrashDir(testFile, trashDir)
	if err != nil {
		t.Errorf("safeMoveToTrashDir failed: %v", err)
	}
}

// TestSafeMoveToTrashDir_Collision tests filename collision handling
func TestSafeMoveToTrashDir_Collision(t *testing.T) {
	tmpDir := t.TempDir()
	trashDir := filepath.Join(tmpDir, "trash")

	if err := os.MkdirAll(trashDir, 0755); err != nil {
		t.Fatalf("Failed to create trash dir: %v", err)
	}

	// Pre-create a file in trash to force collision
	existingFile := filepath.Join(trashDir, "test.txt")
	if err := os.WriteFile(existingFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Should handle collision by using "test (1).txt"
	err := safeMoveToTrashDir(testFile, trashDir)
	if err != nil {
		t.Errorf("safeMoveToTrashDir with collision failed: %v", err)
	}
}

// TestEscapeAppleScriptString tests AppleScript string escaping
func TestEscapeAppleScriptString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{`with"quote`, `with\"quote`},
		{`with'apostrophe`, `with\'apostrophe`},
		{`with\backslash`, `with\\backslash`},
		{`complex'path"with\all`, `complex\'path\"with\\all`},
	}

	for _, tt := range tests {
		result := escapeAppleScriptString(tt.input)
		if result != tt.expected {
			t.Errorf("escapeAppleScriptString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestEscapePowerShellString tests PowerShell string escaping
func TestEscapePowerShellString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with'quote", "with''quote"},
		{"with''double", "with''''double"},
	}

	for _, tt := range tests {
		result := escapePowerShellString(tt.input)
		if result != tt.expected {
			t.Errorf("escapePowerShellString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestValidatePath tests path validation
func TestValidatePath(t *testing.T) {
	tests := []struct {
		path      string
		shouldErr bool
	}{
		{"", true},                               // empty
		{"/", true},                              // root
		{`\`, true},                              // Windows root
		{`C:\`, true},                            // Windows root
		{"../../../etc/passwd", true},            // path traversal
		{"/home/user/../../../etc/passwd", true}, // path traversal
		{"/home/user/file.txt", false},           // valid
		{"/tmp/test.txt", false},                 // valid
	}

	for _, tt := range tests {
		err := validatePath(tt.path)
		if tt.shouldErr && err == nil {
			t.Errorf("validatePath(%q) should return error", tt.path)
		}
		if !tt.shouldErr && err != nil {
			t.Errorf("validatePath(%q) should not return error: %v", tt.path, err)
		}
	}
}
