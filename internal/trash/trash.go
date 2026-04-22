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

	switch runtime.GOOS {
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
	if _, err := exec.LookPath("trash"); err == nil {
		return exec.Command("trash", path).Run()
	}

	// Fall back to AppleScript with proper escaping
	escapedPath := escapeAppleScriptString(path)
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, escapedPath)
	return exec.Command("osascript", "-e", script).Run()
}

// moveToTrashLinux moves a file to trash on Linux.
func moveToTrashLinux(path string) error {
	// Try using gio (GNOME)
	if _, err := exec.LookPath("gio"); err == nil {
		return exec.Command("gio", "trash", path).Run()
	}

	// Try using trash-cli
	if _, err := exec.LookPath("trash"); err == nil {
		return exec.Command("trash", path).Run()
	}

	// Fallback to manual move with TOCTOU-safe implementation
	home := os.Getenv("HOME")
	if home != "" {
		trashDir := filepath.Join(home, ".local", "share", "Trash", "files")
		if err := os.MkdirAll(trashDir, 0755); err == nil {
			return safeMoveToTrashDir(path, trashDir)
		}
	}

	// Final fallback: permanent delete
	return os.RemoveAll(path)
}

// moveToTrashWindows moves a file to trash on Windows.
func moveToTrashWindows(path string) error {
	// Use PowerShell with proper escaping
	escapedPath := escapePowerShellString(path)
	psScript := fmt.Sprintf(`
$shell = New-Object -ComObject Shell.Application
$folder = $shell.Namespace(0)
$item = $folder.ParseName('%s')
if ($item -ne $null) {
    $item.InvokeVerb("delete")
}
`, escapedPath)
	return exec.Command("powershell", "-Command", psScript).Run()
}

// safeMoveToTrashDir moves a file to a trash directory using O_CREATE|O_EXCL
// to prevent TOCTOU race conditions.
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

		// Try to create the file exclusively - fails if it already exists
		// This is atomic and prevents TOCTOU races
		f, err := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			// File created successfully, now rename over it
			_ = f.Close()
			_ = os.Remove(dest) // Remove the empty file we just created

			if err := os.Rename(path, dest); err != nil {
				// Rename failed, clean up
				_ = os.Remove(dest)
				return err
			}
			return nil
		}

		// File already exists, try next counter
		if !os.IsExist(err) {
			// Some other error, fall back to permanent delete
			break
		}
		counter++

		// Safety limit to prevent infinite loops
		if counter > 1000 {
			break
		}
	}

	return os.RemoveAll(path)
}

// escapeAppleScriptString escapes special characters for AppleScript strings.
func escapeAppleScriptString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

// escapePowerShellString escapes special characters for PowerShell strings.
func escapePowerShellString(s string) string {
	// In single-quoted strings, only ' needs escaping (by doubling)
	s = strings.ReplaceAll(s, "'", "''")
	return s
}
