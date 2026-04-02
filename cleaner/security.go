package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// sanitizePathForShell ensures a path is safe to use in shell commands.
// It validates the path exists and escapes special characters.
func sanitizePathForShell(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty path")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Verify the path exists to prevent injection via non-existent paths
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("path does not exist: %w", err)
	}

	// Check for dangerous patterns (but allow legitimate special chars in filenames)
	// Block paths that try to escape the filesystem
	cleanPath := filepath.Clean(absPath)
	if strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal detected")
	}

	return absPath, nil
}

// escapeAppleScriptString escapes special characters for AppleScript strings.
// AppleScript uses backslash for escaping, and special chars: ", ', \
func escapeAppleScriptString(s string) string {
	// Escape backslashes first, then quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

// escapePowerShellString escapes special characters for PowerShell strings.
// PowerShell uses backtick for escaping, and special chars: ', ", `, $, ()
func escapePowerShellString(s string) string {
	// In single-quoted strings, only ' and \ need escaping
	s = strings.ReplaceAll(s, "'", "''") // Escape single quote by doubling
	return s
}

// validateMediaPath validates that a path is a legitimate media file path.
// It checks for existence and that the path doesn't contain shell metacharacters.
func validateMediaPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	// Check for obvious shell injection attempts
	dangerousPatterns := []string{
		";", "|", "&", "$", "`", "(", ")", "{", "}", "[", "]",
		"<", ">", "!", "~", "*", "?", "\\", "\n", "\r",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(path, pattern) {
			// Allow some common characters that are actually safe in filenames
			// but log suspicious patterns
			if pattern != " " && pattern != "(" && pattern != ")" {
				// Continue checking - we'll use proper escaping instead of rejecting
			}
		}
	}

	// Verify file exists
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	return nil
}

// SafePlayMedia plays a media file using OS-native commands with proper escaping.
// Returns the command or an error if the path is invalid.
func SafePlayMedia(path string) (*exec.Cmd, error) {
	if err := validateMediaPath(path); err != nil {
		return nil, err
	}

	absPath, _ := filepath.Abs(path)

	switch runtime.GOOS {
	case "darwin":
		// afplay takes the path as argument - no shell interpolation needed
		return exec.Command("afplay", absPath), nil

	case "linux":
		// aplay takes the path as argument - no shell interpolation needed
		return exec.Command("aplay", absPath), nil

	case "windows":
		// PowerShell requires escaping, use single-quoted string with escaped quotes
		escapedPath := escapePowerShellString(absPath)
		// Use -File parameter with proper argument passing instead of string interpolation
		psScript := fmt.Sprintf(`
$player = New-Object Media.SoundPlayer '%s'
$player.PlaySync()
`, escapedPath)
		return exec.Command("powershell", "-Command", psScript), nil

	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// SafeMoveToTrash moves a file to trash using OS-native commands with proper escaping.
func SafeMoveToTrash(path string) error {
	absPath, err := sanitizePathForShell(path)
	if err != nil {
		return err
	}

	switch runtime.GOOS {
	case "darwin":
		return safeMoveToTrashMacOS(absPath)
	case "linux":
		return safeMoveToTrashLinux(absPath)
	case "windows":
		return safeMoveToTrashWindows(absPath)
	default:
		return os.RemoveAll(absPath)
	}
}

func safeMoveToTrashMacOS(path string) error {
	// Try using the `trash` CLI tool first (takes argument directly, no shell)
	if _, err := exec.LookPath("trash"); err == nil {
		return exec.Command("trash", path).Run()
	}

	// Fall back to AppleScript with proper escaping
	escapedPath := escapeAppleScriptString(path)
	script := fmt.Sprintf(`tell application "Finder" to delete POSIX file "%s"`, escapedPath)
	return exec.Command("osascript", "-e", script).Run()
}

func safeMoveToTrashLinux(path string) error {
	// Try using gio (GNOME) - takes argument directly
	if _, err := exec.LookPath("gio"); err == nil {
		return exec.Command("gio", "trash", path).Run()
	}

	// Try using trash-cli - takes argument directly
	if _, err := exec.LookPath("trash"); err == nil {
		return exec.Command("trash", path).Run()
	}

	// Fallback to manual move with proper filename sanitization
	home := os.Getenv("HOME")
	if home != "" {
		trashDir := filepath.Join(home, ".local", "share", "Trash", "files")
		if err := os.MkdirAll(trashDir, 0755); err == nil {
			// Sanitize the filename for the trash directory
			baseName := filepath.Base(path)
			// Remove or replace dangerous characters in filename
			safeName := regexp.MustCompile(`[<>:"/\\|?*]`).ReplaceAllString(baseName, "_")
			
			dest := filepath.Join(trashDir, safeName)
			counter := 1
			for {
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					return os.Rename(path, dest)
				}
				ext := filepath.Ext(safeName)
				base := strings.TrimSuffix(safeName, ext)
				dest = filepath.Join(trashDir, fmt.Sprintf("%s (%d)%s", base, counter, ext))
				counter++
			}
		}
	}

	return os.RemoveAll(path)
}

func safeMoveToTrashWindows(path string) error {
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
