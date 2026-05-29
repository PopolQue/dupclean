package cleaner

import (
	"os"
	"path/filepath"
)

// getSystemTargets returns OS-specific system cache and temp targets.
func GetSystemTargets() []*CleanTarget {
	switch goos {
	case "darwin":
		return getMacOSTargets()
	case "linux":
		return getLinuxTargets()
	case "windows":
		return getWindowsTargets()
	default:
		return nil
	}
}

func getMacOSTargets() []*CleanTarget {
	home, _ := userHomeDir()
	tmpDir := os.TempDir()

	return []*CleanTarget{
		{
			ID:          "macos-user-cache",
			Category:    "System",
			Label:       "User application cache",
			Description: "Cache files created by your applications",
			Paths:       []string{filepath.Join(home, "Library", "Caches")},
			Patterns:    []string{"*"},
			Risk:        RiskLow,
			OS:          "darwin",
		},
		{
			ID:          "macos-logs-user",
			Category:    "System",
			Label:       "User log files",
			Description: "Log files from your applications",
			Paths:       []string{filepath.Join(home, "Library", "Logs")},
			Patterns:    []string{"*.log", "*.crash"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
		{
			ID:          "macos-tmp",
			Category:    "System",
			Label:       "Temporary files",
			Description: "System and user temporary files",
			Paths:       []string{tmpDir, "/tmp"},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
		{
			ID:          "macos-quicklook",
			Category:    "System",
			Label:       "QuickLook cache",
			Description: "Thumbnail cache for QuickLook previews",
			Paths:       []string{filepath.Join(home, "Library", "Caches", "com.apple.QuickLook.thumbnailcache")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
	}
}

func getLinuxTargets() []*CleanTarget {
	home, _ := userHomeDir()
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache == "" {
		xdgCache = filepath.Join(home, ".cache")
	}

	return []*CleanTarget{
		{
			ID:          "linux-user-cache",
			Category:    "System",
			Label:       "User cache (XDG)",
			Description: "User application cache following XDG specification",
			Paths:       []string{xdgCache},
			Patterns:    []string{"*"},
			Risk:        RiskLow,
			OS:          "linux",
		},
		{
			ID:          "linux-tmp",
			Category:    "System",
			Label:       "System temp",
			Description: "Temporary files in /tmp",
			Paths:       []string{"/tmp"},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "linux",
		},
		{
			ID:          "linux-thumbnails",
			Category:    "System",
			Label:       "Thumbnail cache",
			Description: "Cached thumbnails for file manager previews",
			Paths:       []string{filepath.Join(xdgCache, "thumbnails")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "linux",
		},
	}
}

func getWindowsTargets() []*CleanTarget {
	_ = os.Getenv("USERPROFILE")
	localAppData := os.Getenv("LOCALAPPDATA")
	windir := os.Getenv("WINDIR")

	return []*CleanTarget{
		{
			ID:          "win-user-temp",
			Category:    "System",
			Label:       "User temp folder",
			Description: "Temporary files created by your applications",
			Paths:       []string{filepath.Join(localAppData, "Temp")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "windows",
		},
		{
			ID:          "win-thumbnails",
			Category:    "System",
			Label:       "Thumbnail cache",
			Description: "Cached thumbnails for File Explorer",
			Paths:       []string{filepath.Join(localAppData, "Microsoft", "Windows", "Explorer")},
			Patterns:    []string{"thumbcache_*.db", "iconcache_*.db"},
			Risk:        RiskSafe,
			OS:          "windows",
		},
		{
			ID:          "win-prefetch",
			Category:    "System",
			Label:       "Prefetch cache",
			Description: "Windows prefetch files for faster app startup",
			Paths:       []string{filepath.Join(windir, "Prefetch")},
			Patterns:    []string{"*.pf"},
			Risk:        RiskLow,
			OS:          "windows",
		},
	}
}
