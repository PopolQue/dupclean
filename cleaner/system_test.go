package cleaner

import (
	"os"
	"strings"
	"testing"
)

func TestGetSystemTargetsCrossPlatform(t *testing.T) {
	targets := GetSystemTargets()

	if targets == nil {
		t.Fatal("GetSystemTargets() returned nil")
	}

	if len(targets) == 0 {
		t.Error("GetSystemTargets() returned no targets")
	}

	for _, target := range targets {
		if target.ID == "" {
			t.Error("System target missing ID")
		}
		if target.Category != "System" {
			t.Errorf("Expected category 'System', got %q", target.Category)
		}
		if target.Label == "" {
			t.Errorf("System target %s missing label", target.ID)
		}
	}
}

func TestGetMacOSTargets_Platform(t *testing.T) {
	originalHome, _ := os.UserHomeDir()
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	targets := getMacOSTargets()

	if len(targets) == 0 {
		t.Error("Expected at least 1 macOS system target")
	}

	for _, target := range targets {
		if target.OS != "darwin" {
			t.Errorf("Expected OS 'darwin', got %q", target.OS)
		}
	}
}

func TestGetLinuxTargets_Platform(t *testing.T) {
	targets := getLinuxTargets()

	if len(targets) == 0 {
		t.Error("Expected at least 1 Linux system target")
	}

	for _, target := range targets {
		if target.OS != "linux" {
			t.Errorf("Expected OS 'linux', got %q", target.OS)
		}
	}

	journalFound := false
	thumbnailFound := false
	for _, target := range targets {
		if target.ID == "linux-tmp" {
			journalFound = true
		}
		if target.ID == "linux-thumbnails" {
			thumbnailFound = true
		}
	}
	if !journalFound {
		t.Error("Missing tmp target")
	}
	if !thumbnailFound {
		t.Error("Missing thumbnails target")
	}
}

func TestGetWindowsTargets_Platform(t *testing.T) {
	originalAppData := os.Getenv("LOCALAPPDATA")
	originalTemp := os.Getenv("TEMP")
	defer func() {
		if originalAppData != "" {
			os.Setenv("LOCALAPPDATA", originalAppData)
		}
		if originalTemp != "" {
			os.Setenv("TEMP", originalTemp)
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("LOCALAPPDATA", tmpDir)
	os.Setenv("TEMP", tmpDir)

	targets := getWindowsTargets()

	if len(targets) == 0 {
		t.Error("Expected at least 1 Windows system target")
	}

	for _, target := range targets {
		if target.OS != "windows" {
			t.Errorf("Expected OS 'windows', got %q", target.OS)
		}
	}

	tempFound := false
	for _, target := range targets {
		if target.ID == "win-user-temp" {
			tempFound = true
			if len(target.Paths) > 0 && !strings.Contains(target.Paths[0], "Temp") {
				t.Errorf("Temp files path should contain 'Temp', got %q", target.Paths[0])
			}
		}
	}
	if !tempFound {
		t.Error("Missing temp files target")
	}
}
