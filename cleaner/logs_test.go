package cleaner

import (
	"os"
	"testing"
)

func TestGetLogsTargets_CrossPlatform(t *testing.T) {
	oldOS := goos
	defer func() { goos = oldOS }()

	platforms := []string{"darwin", "linux", "windows", "unsupported"}

	for _, p := range platforms {
		t.Run("Platform_"+p, func(t *testing.T) {
			goos = p
			targets := GetLogsTargets()

			if p == "unsupported" {
				if targets != nil {
					t.Errorf("Expected nil for unsupported platform")
				}
				return
			}

			if targets == nil {
				t.Fatalf("GetLogsTargets() returned nil for %s", p)
			}
			if len(targets) == 0 {
				t.Errorf("GetLogsTargets() returned 0 targets for %s", p)
			}
		})
	}
}

func TestGetLogsTargetsMac_Logic(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	targets := getLogsTargetsMac()

	if len(targets) == 0 {
		t.Error("Expected at least 1 macOS logs target")
	}

	systemLogsFound := false
	for _, target := range targets {
		if target.ID == "macos-crash-reports" || target.ID == "macos-app-logs" {
			systemLogsFound = true
		}
	}
	if !systemLogsFound {
		t.Error("Missing macOS logs target")
	}
}

func TestGetLogsTargetsLinux_Logic(t *testing.T) {
	targets := getLogsTargetsLinux()

	if len(targets) == 0 {
		t.Error("Expected at least 1 Linux logs target")
	}

	systemLogsFound := false
	for _, target := range targets {
		if target.ID == "linux-old-logs" {
			systemLogsFound = true
		}
	}
	if !systemLogsFound {
		t.Error("Missing Linux logs target")
	}
}

func TestGetLogsTargetsWindows_Logic(t *testing.T) {
	targets := getLogsTargetsWindows()

	if len(targets) == 0 {
		t.Error("Expected at least 1 Windows logs target")
	}

	windowsLogsFound := false
	for _, t := range targets {
		if t.OS == "windows" {
			windowsLogsFound = true
		}
	}
	if !windowsLogsFound {
		t.Error("Missing Windows logs target")
	}
}
