package cleaner

import (
	"os"
	"strings"
	"testing"
)

func TestGetSystemTargets_CrossPlatform(t *testing.T) {
	oldOS := goos
	defer func() { goos = oldOS }()

	platforms := []string{"darwin", "linux", "windows", "unsupported"}

	for _, p := range platforms {
		t.Run("Platform_"+p, func(t *testing.T) {
			goos = p
			targets := GetSystemTargets()

			if p == "unsupported" {
				if targets != nil {
					t.Errorf("Expected nil for unsupported platform")
				}
				return
			}

			if targets == nil {
				t.Fatalf("GetSystemTargets() returned nil for %s", p)
			}
			if len(targets) == 0 {
				t.Errorf("GetSystemTargets() returned 0 targets for %s", p)
			}
		})
	}
}

func TestGetMacOSTargets_Logic(t *testing.T) {
	originalHome := os.Getenv("HOME")
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

func TestGetLinuxTargets_Logic(t *testing.T) {
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

func TestGetWindowsTargets_Logic(t *testing.T) {
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
