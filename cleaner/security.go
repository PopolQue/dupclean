package cleaner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"dupclean/internal/trash"
)

// SafePlayMedia plays a media file using OS-native commands with proper escaping.
func SafePlayMedia(path string) (*exec.Cmd, error) {
	if err := validateMediaPath(path); err != nil {
		return nil, err
	}

	absPath, _ := filepath.Abs(path)

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("afplay", absPath), nil
	case "linux":
		return exec.Command("aplay", absPath), nil
	case "windows":
		escapedPath := escapePowerShellString(absPath)
		psScript := `
$player = New-Object Media.SoundPlayer '` + escapedPath + `'
$player.PlaySync()
`
		return exec.Command("powershell", "-Command", psScript), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// SafeMoveToTrash moves a file to trash using the unified internal/trash package.
// This is a wrapper for backwards compatibility.
func SafeMoveToTrash(path string) error {
	return trash.MoveToTrash(path)
}

// escapePowerShellString escapes special characters for PowerShell strings.
func escapePowerShellString(s string) string {
	// In single-quoted strings, only ' needs escaping (by doubling)
	return strings.ReplaceAll(s, "'", "''")
}

// validateMediaPath validates that a path is a legitimate media file path.
func validateMediaPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}
	return nil
}
