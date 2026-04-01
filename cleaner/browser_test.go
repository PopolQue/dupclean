package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetBrowserTargetsCrossPlatform(t *testing.T) {
	targets := GetBrowserTargets()

	if targets == nil {
		t.Fatal("GetBrowserTargets() returned nil")
	}

	// Should return targets for current OS
	if len(targets) == 0 {
		t.Skip("No browser targets for this OS (expected in some environments)")
	}

	// Verify all targets have required fields
	for _, target := range targets {
		if target.ID == "" {
			t.Error("Browser target missing ID")
		}
		if target.Category != "Browser" {
			t.Errorf("Expected category 'Browser', got %q", target.Category)
		}
		if target.Label == "" {
			t.Errorf("Browser target %s missing label", target.ID)
		}
		if len(target.Paths) == 0 {
			t.Errorf("Browser target %s has no paths", target.ID)
		}
	}
}

func TestGetBrowserTargetsMac_Platform(t *testing.T) {
	originalHome, _ := os.UserHomeDir()
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	targets := getBrowserTargetsMac()

	if len(targets) != 3 {
		t.Errorf("Expected 3 macOS browser targets, got %d", len(targets))
	}

	safariFound := false
	for _, target := range targets {
		if target.ID == "macos-safari-cache" {
			safariFound = true
			expectedPath := filepath.Join(tmpDir, "Library", "Caches", "com.apple.Safari")
			if len(target.Paths) > 0 && target.Paths[0] != expectedPath {
				t.Errorf("Safari path = %q, want %q", target.Paths[0], expectedPath)
			}
		}
	}
	if !safariFound {
		t.Error("Missing Safari cache target")
	}
}

func TestGetBrowserTargetsLinux_Platform(t *testing.T) {
	originalHome, _ := os.UserHomeDir()
	defer func() {
		if originalHome != "" {
			os.Setenv("HOME", originalHome)
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	targets := getBrowserTargetsLinux()

	if len(targets) != 2 {
		t.Errorf("Expected 2 Linux browser targets, got %d", len(targets))
	}

	chromeFound := false
	for _, target := range targets {
		if target.ID == "linux-chrome-cache" {
			chromeFound = true
			expectedPath := filepath.Join(tmpDir, ".config", "google-chrome", "Default", "Cache")
			if len(target.Paths) > 0 && target.Paths[0] != expectedPath {
				t.Errorf("Chrome path = %q, want %q", target.Paths[0], expectedPath)
			}
		}
	}
	if !chromeFound {
		t.Error("Missing Chrome cache target")
	}
}

func TestGetBrowserTargetsWindows_Platform(t *testing.T) {
	originalAppData := os.Getenv("LOCALAPPDATA")
	defer func() {
		if originalAppData != "" {
			os.Setenv("LOCALAPPDATA", originalAppData)
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("LOCALAPPDATA", tmpDir)

	targets := getBrowserTargetsWindows()

	if len(targets) != 2 {
		t.Errorf("Expected 2 Windows browser targets, got %d", len(targets))
	}

	chromeFound := false
	edgeFound := false
	for _, target := range targets {
		if target.ID == "win-chrome-cache" {
			chromeFound = true
			expectedPath := filepath.Join(tmpDir, "Google", "Chrome", "User Data", "Default", "Cache")
			if len(target.Paths) > 0 && target.Paths[0] != expectedPath {
				t.Errorf("Chrome path = %q, want %q", target.Paths[0], expectedPath)
			}
		}
		if target.ID == "win-edge-cache" {
			edgeFound = true
		}
	}
	if !chromeFound {
		t.Error("Missing Chrome cache target")
	}
	if !edgeFound {
		t.Error("Missing Edge cache target")
	}
}
