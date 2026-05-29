package cleaner

import (
	"os"
	"testing"
)

func TestGetDeveloperTargets_CrossPlatform(t *testing.T) {
	oldOS := goos
	defer func() { goos = oldOS }()

	platforms := []string{"darwin", "linux", "windows", "unsupported"}

	for _, p := range platforms {
		t.Run("Platform_"+p, func(t *testing.T) {
			goos = p
			targets := GetDeveloperTargets()

			if p == "unsupported" {
				if targets != nil {
					t.Errorf("Expected nil for unsupported platform, got %d targets", len(targets))
				}
				return
			}

			if targets == nil {
				t.Fatalf("GetDeveloperTargets() returned nil for %s", p)
			}
			if len(targets) == 0 {
				t.Errorf("GetDeveloperTargets() returned 0 targets for %s", p)
			}
		})
	}
}

func TestGetDeveloperTargetsMacOS_Logic(t *testing.T) {
	oldHomeDir := userHomeDir
	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	defer func() { userHomeDir = oldHomeDir }()

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

func TestGetDeveloperTargetsLinux_Logic(t *testing.T) {
	oldHomeDir := userHomeDir
	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }
	defer func() { userHomeDir = oldHomeDir }()

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

func TestGetDeveloperTargetsWindows_Logic(t *testing.T) {
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
