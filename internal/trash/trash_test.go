package trash

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// mock setup utilities
func setupMockExec(lookPathFunc func(string) (string, error), cmdFunc func(string, ...string) *exec.Cmd, os string) func() {
	oldLookPath := execLookPath
	oldCmd := execCommand
	oldOS := goos

	if lookPathFunc != nil {
		execLookPath = lookPathFunc
	}
	if cmdFunc != nil {
		execCommand = cmdFunc
	}
	if os != "" {
		goos = os
	}

	return func() {
		execLookPath = oldLookPath
		execCommand = oldCmd
		goos = oldOS
	}
}

// helper command that immediately returns success when run
func mockSuccessCmd(name string, args ...string) *exec.Cmd {
	return exec.Command("true")
}

// helper command that fails when run
func mockFailCmd(name string, args ...string) *exec.Cmd {
	return exec.Command("false")
}

// TestMoveToTrash_CrossPlatform tests all OS branches of MoveToTrash
func TestMoveToTrash_CrossPlatform(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	platforms := []string{"darwin", "linux", "windows", "freebsd"}

	for _, p := range platforms {
		t.Run("Platform_"+p, func(t *testing.T) {
			cleanup := setupMockExec(nil, mockSuccessCmd, p)
			defer cleanup()

			err := MoveToTrash(testFile)
			if err != nil {
				t.Errorf("MoveToTrash failed for %s: %v", p, err)
			}
		})
	}
}

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
	t.Logf("MoveToTrash non-existent file: %v", err)
}

// --- macOS Tests ---

func TestMoveToTrashMacOS_TrashCliSuccess(t *testing.T) {
	cleanup := setupMockExec(
		func(file string) (string, error) {
			if file == "trash" {
				return "/usr/bin/trash", nil
			}
			return "", errors.New("not found")
		},
		mockSuccessCmd,
		"darwin",
	)
	defer cleanup()

	err := moveToTrashMacOS("/dummy/path")
	if err != nil {
		t.Errorf("Expected success when trash cli is found, got %v", err)
	}
}

func TestMoveToTrashMacOS_AppleScriptSuccess(t *testing.T) {
	cleanup := setupMockExec(
		func(file string) (string, error) {
			return "", errors.New("not found") // trash cli missing
		},
		mockSuccessCmd,
		"darwin",
	)
	defer cleanup()

	err := moveToTrashMacOS("/dummy/path")
	if err != nil {
		t.Errorf("Expected success on AppleScript fallback, got %v", err)
	}
}

// --- Linux Tests ---

func TestMoveToTrashLinux_GioSuccess(t *testing.T) {
	cleanup := setupMockExec(
		func(file string) (string, error) {
			if file == "gio" {
				return "/usr/bin/gio", nil
			}
			return "", errors.New("not found")
		},
		mockSuccessCmd,
		"linux",
	)
	defer cleanup()

	err := moveToTrashLinux("/dummy/path")
	if err != nil {
		t.Errorf("Expected success when gio is found, got %v", err)
	}
}

func TestMoveToTrashLinux_TrashCliSuccess(t *testing.T) {
	cleanup := setupMockExec(
		func(file string) (string, error) {
			if file == "trash" {
				return "/usr/bin/trash", nil
			}
			return "", errors.New("not found")
		},
		mockSuccessCmd,
		"linux",
	)
	defer cleanup()

	err := moveToTrashLinux("/dummy/path")
	if err != nil {
		t.Errorf("Expected success when trash is found, got %v", err)
	}
}

func TestMoveToTrashLinux_FallbackToSafeMove(t *testing.T) {
	// Mock everything missing so it falls back to HOME/.local/share/Trash
	cleanup := setupMockExec(
		func(file string) (string, error) { return "", errors.New("not found") },
		nil,
		"linux",
	)
	defer cleanup()

	// Set up dummy home and file
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")

	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	err := moveToTrashLinux(testFile)
	if err != nil {
		t.Errorf("Fallback to safeMoveToTrashDir failed: %v", err)
	}

	// Verify it was moved
	trashDir := filepath.Join(tmpDir, ".local", "share", "Trash", "files")
	trashedFile := filepath.Join(trashDir, "test.txt")
	if _, err := os.Stat(trashedFile); os.IsNotExist(err) {
		t.Errorf("File was not moved to the fallback trash directory")
	}
}

func TestMoveToTrashLinux_FallbackToPermanent(t *testing.T) {
	cleanup := setupMockExec(
		func(file string) (string, error) { return "", errors.New("not found") },
		nil,
		"linux",
	)
	defer cleanup()

	// Unset HOME so it skips the safeMoveToTrashDir block
	os.Unsetenv("HOME")

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	err := moveToTrashLinux(testFile)
	if err != nil {
		t.Errorf("Fallback to os.RemoveAll failed: %v", err)
	}

	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Errorf("File was not permanently deleted")
	}
}

// --- Windows Tests ---

func TestMoveToTrashWindows_Success(t *testing.T) {
	cleanup := setupMockExec(nil, mockSuccessCmd, "windows")
	defer cleanup()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(testFile, []byte("test"), 0644)

	err := moveToTrashWindows(testFile)
	if err != nil {
		t.Errorf("Expected success on PowerShell command, got %v", err)
	}
}

// --- Other Utility Tests ---

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
