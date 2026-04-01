package cleaner

import (
	"os"
	"runtime"
	"testing"
)

func TestGetDeveloperTargetsCrossPlatform(t *testing.T) {
	targets := GetDeveloperTargets()

	if targets == nil {
		t.Fatal("GetDeveloperTargets() returned nil")
	}

	if len(targets) == 0 {
		t.Error("GetDeveloperTargets() returned no targets")
	}

	for _, target := range targets {
		if target.ID == "" {
			t.Error("Developer target missing ID")
		}
		if target.Category != "Developer" {
			t.Errorf("Expected category 'Developer', got %q", target.Category)
		}
		if target.Label == "" {
			t.Errorf("Developer target %s missing label", target.ID)
		}
	}

	goCacheFound := false
	for _, t := range targets {
		if t.ID == "dev-go-cache" {
			goCacheFound = true
		}
	}
	if !goCacheFound {
		t.Error("Missing Go cache target")
	}
}

func TestGetDeveloperTargetsMacOS_Platform(t *testing.T) {
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

	targets := getDeveloperTargetsMacOS()

	if len(targets) == 0 {
		t.Error("Expected at least 1 macOS developer target")
	}

	cocoaPodsFound := false
	for _, target := range targets {
		if target.ID == "dev-xcode-derived" || target.ID == "dev-npm-cache" {
			cocoaPodsFound = true
		}
	}
	if !cocoaPodsFound {
		t.Error("Missing Xcode or npm target")
	}
}

func TestGetDeveloperTargetsLinux_Platform(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux test on " + runtime.GOOS)
	}
	targets := getDeveloperTargetsLinux()

	if len(targets) == 0 {
		t.Error("Expected at least 1 Linux developer target")
	}

	for _, target := range targets {
		if target.OS != "linux" {
			t.Errorf("Expected OS 'linux', got %q", target.OS)
		}
	}
}

func TestGetDeveloperTargetsWindows_Platform(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows test on " + runtime.GOOS)
	}
	originalAppData := os.Getenv("LOCALAPPDATA")
	originalUserProfile := os.Getenv("USERPROFILE")
	defer func() {
		if originalAppData != "" {
			os.Setenv("LOCALAPPDATA", originalAppData)
		}
		if originalUserProfile != "" {
			os.Setenv("USERPROFILE", originalUserProfile)
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("LOCALAPPDATA", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)

	targets := getDeveloperTargetsWindows()

	if len(targets) == 0 {
		t.Error("Expected at least 1 Windows developer target")
	}

	for _, target := range targets {
		if target.OS != "windows" {
			t.Errorf("Expected OS 'windows', got %q", target.OS)
		}
	}
}

func TestGetGoCachePath(t *testing.T) {
	originalGoCache := os.Getenv("GOCACHE")
	defer func() {
		if originalGoCache != "" {
			os.Setenv("GOCACHE", originalGoCache)
		} else {
			os.Unsetenv("GOCACHE")
		}
	}()

	tmpDir := t.TempDir()
	os.Setenv("GOCACHE", tmpDir)

	path := getGoCachePath()
	if path != tmpDir {
		t.Errorf("getGoCachePath() = %q, want %q", path, tmpDir)
	}

	os.Unsetenv("GOCACHE")
	path = getGoCachePath()
	if path == "" {
		t.Error("getGoCachePath() returned empty string")
	}
}
