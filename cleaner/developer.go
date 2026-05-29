package cleaner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetDeveloperTargets returns developer tool cache targets.
func GetDeveloperTargets() []*CleanTarget {
	switch goos {
	case "darwin":
		return getDeveloperTargetsMacOS()
	case "linux":
		return getDeveloperTargetsLinux()
	case "windows":
		return getDeveloperTargetsWindows()
	default:
		return nil
	}
}

func getDeveloperTargetsMacOS() []*CleanTarget {
	home, _ := userHomeDir()

	var targets []*CleanTarget

	// Xcode DerivedData - safe to delete
	targets = append(targets, &CleanTarget{
		ID:          "dev-xcode-derived",
		Category:    "Developer",
		Label:       "Xcode DerivedData",
		Description: "Xcode build intermediates (safe to delete)",
		Paths:       []string{filepath.Join(home, "Library", "Developer", "Xcode", "DerivedData")},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
		OS:          "darwin",
	})

	// Check if Go is installed and get cache paths
	if goCache := getGoCachePath(); goCache != "" {
		targets = append(targets, &CleanTarget{
			ID:          "dev-go-cache",
			Category:    "Developer",
			Label:       "Go build cache",
			Description: "Go compiler cache (will rebuild as needed)",
			Paths:       []string{goCache},
			Patterns:    []string{"*"},
			Risk:        RiskLow,
			OS:          "darwin",
		})
	}

	// npm cache
	targets = append(targets, &CleanTarget{
		ID:          "dev-npm-cache",
		Category:    "Developer",
		Label:       "npm cache",
		Description: "npm package cache",
		Paths:       []string{filepath.Join(home, ".npm", "_cacache")},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
		OS:          "darwin",
	})

	// Cargo (Rust) cache
	targets = append(targets, &CleanTarget{
		ID:          "dev-cargo-cache",
		Category:    "Developer",
		Label:       "Rust Cargo cache",
		Description: "Rust Cargo registry cache",
		Paths:       []string{filepath.Join(home, ".cargo", "registry", "cache")},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
		OS:          "darwin",
	})

	return targets
}

func getDeveloperTargetsLinux() []*CleanTarget {
	home, _ := userHomeDir()
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache == "" {
		xdgCache = filepath.Join(home, ".cache")
	}

	var targets []*CleanTarget

	// Go cache
	if goCache := getGoCachePath(); goCache != "" {
		targets = append(targets, &CleanTarget{
			ID:          "dev-go-cache",
			Category:    "Developer",
			Label:       "Go build cache",
			Description: "Go compiler cache (will rebuild as needed)",
			Paths:       []string{goCache},
			Patterns:    []string{"*"},
			Risk:        RiskLow,
			OS:          "linux",
		})
	}

	// npm cache
	targets = append(targets, &CleanTarget{
		ID:          "dev-npm-cache",
		Category:    "Developer",
		Label:       "npm cache",
		Description: "npm package cache",
		Paths:       []string{filepath.Join(home, ".npm", "_cacache")},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
		OS:          "linux",
	})

	// pip cache
	targets = append(targets, &CleanTarget{
		ID:          "dev-pip-cache",
		Category:    "Developer",
		Label:       "pip cache",
		Description: "Python pip package cache",
		Paths:       []string{filepath.Join(xdgCache, "pip")},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
		OS:          "linux",
	})

	// Cargo cache
	targets = append(targets, &CleanTarget{
		ID:          "dev-cargo-cache",
		Category:    "Developer",
		Label:       "Rust Cargo cache",
		Description: "Rust Cargo registry cache",
		Paths:       []string{filepath.Join(home, ".cargo", "registry", "cache")},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
		OS:          "linux",
	})

	return targets
}

func getDeveloperTargetsWindows() []*CleanTarget {
	localAppData := os.Getenv("LOCALAPPDATA")

	var targets []*CleanTarget

	// npm cache
	targets = append(targets, &CleanTarget{
		ID:          "dev-npm-cache",
		Category:    "Developer",
		Label:       "npm cache",
		Description: "npm package cache",
		Paths:       []string{filepath.Join(localAppData, "npm-cache")},
		Patterns:    []string{"*"},
		Risk:        RiskSafe,
		OS:          "windows",
	})

	return targets
}

// getGoCachePath returns the Go cache directory by running `go env GOCACHE`.
func getGoCachePath() string {
	cmd := exec.Command("go", "env", "GOCACHE")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
