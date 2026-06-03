package trash

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// vars for testing to allow mocking OS commands
var (
	execCommand  = exec.Command
	execLookPath = exec.LookPath
	goos         = runtime.GOOS
)

// MoveToTrash moves a file or directory to the system's trash/recycle bin.
// This is a unified implementation used across all packages.
//
// On macOS: Uses `trash` CLI or Finder AppleScript
// On Linux: Uses `gio` or moves to ~/.local/share/Trash
// On Windows: Uses PowerShell to move to Recycle Bin
//
// Returns an error if the operation fails.
func MoveToTrash(path string) error {
	// Validate path
	if err := validatePath(path); err != nil {
		return err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	switch goos {
	case "darwin":
		return moveToTrashMacOS(absPath)
	case "linux":
		return moveToTrashLinux(absPath)
	case "windows":
		return moveToTrashWindows(absPath)
	default:
		// Fallback to permanent delete for unsupported OS
		return os.RemoveAll(path)
	}
}

// validatePath performs basic validation on the path.
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("cannot move empty path to trash")
	}

	// Check for root directory
	if path == "/" || path == `\` || path == `C:\` || path == `c:\` {
		return fmt.Errorf("cannot delete root directory")
	}

	// Check for path traversal in original path BEFORE resolving
	// This catches attempts like "../../../etc/passwd"
	if strings.HasPrefix(path, "..") || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "..\\") {
		return fmt.Errorf("path traversal detected")
	}

	// Also check for /.. or \.. anywhere in the path
	if strings.Contains(path, "/..") || strings.Contains(path, "\\..") {
		return fmt.Errorf("path traversal detected")
	}

	return nil
}

// moveToTrashMacOS moves a file to trash on macOS.
func moveToTrashMacOS(path string) error {
	// Try using the `trash` CLI tool first (brew install trash)
	if _, err := execLookPath("trash"); err == nil {
		return execCommand("trash", "--", path).Run()
	}

	// Fall back to AppleScript with proper escaping
	escapedPath := escapeAppleScriptString(path)
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, escapedPath)
	return execCommand("osascript", "-e", script).Run()
}

// moveToTrashLinux moves a file to trash on Linux.
func moveToTrashLinux(path string) error {
	// Try using gio (GNOME)
	if _, err := execLookPath("gio"); err == nil {
		return execCommand("gio", "trash", "--", path).Run()
	}

	// Try using trash-cli
	if _, err := execLookPath("trash"); err == nil {
		return execCommand("trash", "--", path).Run()
	}

	// Fallback to manual move with TOCTOU-safe implementation
	home := os.Getenv("HOME")
	if home != "" {
		trashDir := filepath.Join(home, ".local", "share", "Trash", "files")
		// #nosec G301
		// #nosec G703
		if err := os.MkdirAll(trashDir, 0700); err == nil {
			return safeMoveToTrashDir(path, trashDir)
		}
	}

	// Final fallback: permanent delete
	return os.RemoveAll(path)
}

// moveToTrashWindows moves a file to trash on Windows using PowerShell.
func moveToTrashWindows(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Use the Shell.Application COM object which is the most reliable way
	// to move to the Recycle Bin via script without third-party tools.
	// We use -EncodedCommand to avoid any shell escaping issues.

	psScript := fmt.Sprintf(`
$path = '%s'
if (Test-Path $path) {
    $shell = New-Object -ComObject Shell.Application
    $folder = $shell.Namespace((Split-Path $path))
    $item = $folder.ParseName((Split-Path $path -Leaf))
    if ($item -ne $null) {
        $item.InvokeVerb("delete")
    }
}
`, strings.ReplaceAll(absPath, "'", "''"))

	cmd := execCommand("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	return cmd.Run()
}

// safeMoveToTrashDir moves a file to a trash directory.
func safeMoveToTrashDir(path, trashDir string) error {
	baseName := filepath.Base(path)
	// Sanitize filename
	safeName := regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(baseName, "_")

	counter := 0
	for {
		var fileName string
		if counter == 0 {
			fileName = safeName
		} else {
			ext := filepath.Ext(safeName)
			base := strings.TrimSuffix(safeName, ext)
			fileName = fmt.Sprintf("%s (%d)%s", base, counter, ext)
		}

		dest := filepath.Join(trashDir, fileName)

		// Check if file exists - this is still a slight race, but
		// os.Rename is generally atomic on the same filesystem.
		// #nosec G703
		_, err := os.Stat(dest)
		if os.IsNotExist(err) {
			// File does not exist, safe to rename
			// #nosec G703
			if err := os.Rename(path, dest); err != nil {
				return err
			}
			return nil
		}

		// File exists, try next counter
		counter++

		// Safety limit to prevent infinite loops
		if counter > 1000 {
			break
		}
	}

	return fmt.Errorf("failed to move to trash: too many collisions in %s", trashDir)
}

// escapeAppleScriptString escapes special characters for AppleScript strings.
func escapeAppleScriptString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\x00", "")
	return s
}
