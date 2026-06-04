package cleaner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// SafePlayMedia plays a media file using OS-native commands with proper escaping.
func SafePlayMedia(ctx context.Context, path string) (*exec.Cmd, error) {
	if err := validateMediaPath(path); err != nil {
		return nil, err
	}

	absPathVal, _ := absPath(path)

	switch goos {
	case "darwin":
		return execCommandContext(ctx, "afplay", absPathVal), nil
	case "linux":
		return execCommandContext(ctx, "aplay", absPathVal), nil
	case "windows":
		psScript := fmt.Sprintf(`
$path = '%s'
$player = New-Object Media.SoundPlayer $path
$player.PlaySync()
`, strings.ReplaceAll(absPathVal, "'", "''"))
		return execCommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", psScript), nil
	default:
		return nil, fmt.Errorf("unsupported OS: %s", goos)
	}
}

// SafeMoveToTrash moves a file to trash using the unified internal/trash package.
// This is a wrapper for backwards compatibility.
func SafeMoveToTrash(path string) error {
	return moveToTrash(path)
}

// validateMediaPath validates that a path is a legitimate media file path.
func validateMediaPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	absPathVal, err := absPath(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}
	if _, err := os.Stat(absPathVal); err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}
	return nil
}
