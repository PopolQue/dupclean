package cleaner

import (
	"os"
	"path/filepath"
)

// GetLogsTargets returns log file targets for the current OS.
func GetLogsTargets() []*CleanTarget {
	switch goos {
	case "darwin":
		return getLogsTargetsMac()
	case "linux":
		return getLogsTargetsLinux()
	case "windows":
		return getLogsTargetsWindows()
	default:
		return nil
	}
}


func getLogsTargetsMac() []*CleanTarget {
	home, _ := os.UserHomeDir()
	lib := filepath.Join(home, "Library")

	return []*CleanTarget{
		{
			ID:          "macos-crash-reports",
			Category:    "Logs",
			Label:       "Crash reports",
			Description: "Application crash reports and diagnostics",
			Paths:       []string{filepath.Join(lib, "Logs", "DiagnosticReports")},
			Patterns:    []string{"*.crash", "*.ips"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
		{
			ID:          "macos-app-logs",
			Category:    "Logs",
			Label:       "App log files",
			Description: "Old application log files",
			Paths:       []string{filepath.Join(lib, "Logs")},
			Patterns:    []string{"*.log"},
			Risk:        RiskSafe,
			OS:          "darwin",
		},
	}
}

func getLogsTargetsLinux() []*CleanTarget {
	return []*CleanTarget{
		{
			ID:          "linux-old-logs",
			Category:    "Logs",
			Label:       "Rotated log files",
			Description: "Old rotated and compressed log files",
			Paths:       []string{"/var/log"},
			Patterns:    []string{"*.gz", "*.[0-9]"},
			Risk:        RiskLow,
			OS:          "linux",
		},
	}
}

func getLogsTargetsWindows() []*CleanTarget {
	localAppData := os.Getenv("LOCALAPPDATA")

	return []*CleanTarget{
		{
			ID:          "win-error-reports",
			Category:    "Logs",
			Label:       "Error reports",
			Description: "Windows Error Reporting files",
			Paths:       []string{filepath.Join(localAppData, "Microsoft", "Windows", "WER")},
			Patterns:    []string{"*"},
			Risk:        RiskSafe,
			OS:          "windows",
		},
	}
}
