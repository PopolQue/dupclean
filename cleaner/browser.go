package cleaner

import (
	"os"
	"path/filepath"
)

// GetBrowserTargets returns browser cache targets for the current OS.
func GetBrowserTargets() []*CleanTarget {
	switch goos {
	case "darwin":
		return getBrowserTargetsMac()
	case "linux":
		return getBrowserTargetsLinux()
	case "windows":
		return getBrowserTargetsWindows()
	default:
		return nil
	}
}


func getBrowserTargetsMac() []*CleanTarget {
	home, _ := os.UserHomeDir()
	lib := filepath.Join(home, "Library")

	return []*CleanTarget{
		{
			ID:          "macos-safari-cache",
			Category:    "Browser",
			Label:       "Safari cache",
			Description: "Safari browser cache and website data",
			Paths:       []string{filepath.Join(lib, "Caches", "com.apple.Safari")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
		{
			ID:          "macos-chrome-cache",
			Category:    "Browser",
			Label:       "Chrome cache",
			Description: "Google Chrome browser cache",
			Paths:       []string{filepath.Join(lib, "Application Support", "Google", "Chrome", "Default", "Cache")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
		{
			ID:          "macos-firefox-cache",
			Category:    "Browser",
			Label:       "Firefox cache",
			Description: "Mozilla Firefox browser cache",
			Paths:       []string{filepath.Join(lib, "Caches", "Firefox")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
	}
}

func getBrowserTargetsLinux() []*CleanTarget {
	home, _ := os.UserHomeDir()
	config := filepath.Join(home, ".config")

	return []*CleanTarget{
		{
			ID:          "linux-chrome-cache",
			Category:    "Browser",
			Label:       "Chrome cache",
			Description: "Google Chrome browser cache",
			Paths:       []string{filepath.Join(config, "google-chrome", "Default", "Cache")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "linux",
		},
		{
			ID:          "linux-firefox-cache",
			Category:    "Browser",
			Label:       "Firefox cache",
			Description: "Mozilla Firefox browser cache",
			Paths:       []string{filepath.Join(home, ".mozilla", "firefox")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "linux",
		},
	}
}

func getBrowserTargetsWindows() []*CleanTarget {
	localAppData := os.Getenv("LOCALAPPDATA")

	return []*CleanTarget{
		{
			ID:          "win-chrome-cache",
			Category:    "Browser",
			Label:       "Chrome cache",
			Description: "Google Chrome browser cache",
			Paths:       []string{filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Cache")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "windows",
		},
		{
			ID:          "win-edge-cache",
			Category:    "Browser",
			Label:       "Edge cache",
			Description: "Microsoft Edge browser cache",
			Paths:       []string{filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Cache")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "windows",
		},
	}
}
