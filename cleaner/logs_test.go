package cleaner

import (
	"os"
	"runtime"
	"testing"
)

func TestGetLogsTargetsCrossPlatform(t *testing.T) {
	targets := GetLogsTargets()

	if targets == nil {
		t.Fatal("GetLogsTargets() returned nil")
	}

	if len(targets) == 0 {
		t.Error("GetLogsTargets() returned no targets")
	}

	for _, target := range targets {
		if target.ID == "" {
			t.Error("Logs target missing ID")
		}
		if target.Category != "Logs" {
			t.Errorf("Expected category 'Logs', got %q", target.Category)
		}
		if target.Label == "" {
			t.Errorf("Logs target %s missing label", target.ID)
		}
	}
}

func TestGetLogsTargetsMac_Platform(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS test on " + runtime.GOOS)
	}
	originalHome, _ := os.UserHomeDir()
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

func TestGetLogsTargetsLinux_Platform(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux test on " + runtime.GOOS)
	}
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

func TestGetLogsTargetsWindows_Platform(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows test on " + runtime.GOOS)
	}
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
