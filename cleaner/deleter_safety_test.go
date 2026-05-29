package cleaner

import (
	"errors"
	"os"
	"syscall"
	"testing"
)

func TestVerifyDeletionSafety_CrossPlatform(t *testing.T) {
	oldOS := goos
	oldHomeDir := userHomeDir
	defer func() {
		goos = oldOS
		userHomeDir = oldHomeDir
	}()

	tmpDir := t.TempDir()
	userHomeDir = func() (string, error) { return tmpDir, nil }

	tests := []struct {
		name    string
		path    string
		os      string
		wantErr bool
	}{
		{"Unix root", "/", "linux", true},
		{"Windows root", "C:\\", "windows", true},
		{"Protected path", "/etc", "linux", true},
		{"User home", tmpDir, "linux", true},
		{"Safe path", "/home/user/safe", "linux", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goos = tt.os
			err := verifyDeletionSafety(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyDeletionSafety(%q) on %s error = %v, wantErr %v", tt.path, tt.os, err, tt.wantErr)
			}
		})
	}
}

func TestIsFileInUse_CrossPlatform(t *testing.T) {
	oldOS := goos
	defer func() { goos = oldOS }()

	tests := []struct {
		name     string
		err      error
		os       string
		expected bool
	}{
		{"Nil error", nil, "linux", false},
		{"Permission error", os.ErrPermission, "linux", true},
		{"Busy string", errors.New("file is busy"), "linux", true},
		{"In use string", errors.New("file in use"), "linux", true},
		{"Windows sharing violation", syscall.Errno(32), "windows", true},
		{"Windows access denied", syscall.Errno(5), "windows", true},
		{"Random error", errors.New("random error"), "linux", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goos = tt.os
			result := isFileInUse(tt.err)
			if result != tt.expected {
				t.Errorf("isFileInUse(%v) on %s = %v, want %v", tt.err, tt.os, result, tt.expected)
			}
		})
	}
}

// TestDeleteEntry_SafetyChecks tests that deleteEntry has proper safety checks
func TestDeleteEntry_SafetyChecks(t *testing.T) {
	oldOS := goos
	oldAbs := absPath
	defer func() {
		goos = oldOS
		absPath = oldAbs
	}()

	// Mock absPath to return path as-is for easy root testing
	absPath = func(path string) (string, error) { return path, nil }

	tests := []struct {
		name      string
		path      string
		os        string
		permanent bool
		wantErr   bool
	}{
		{"empty path", "", "linux", true, true},
		{"empty path non-permanent", "", "linux", false, true},
		{"root unix", "/", "linux", true, true},
		{"root windows", `C:\`, "windows", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goos = tt.os
			entry := EntryInfo{
				Path: tt.path,
				Size: 100,
			}

			deleted, freed, skipped, err := deleteEntry(entry, tt.permanent)

			if tt.wantErr && err == nil {
				t.Errorf("deleteEntry(%q) expected error, got nil", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("deleteEntry(%q) unexpected error: %v", tt.path, err)
			}
			if deleted != 0 {
				t.Errorf("deleteEntry(%q) deleted = %d, want 0", tt.path, deleted)
			}
			if freed != 0 {
				t.Errorf("deleteEntry(%q) freed = %d, want 0", tt.path, freed)
			}
			if skipped {
				t.Errorf("deleteEntry(%q) skipped = true, want false", tt.path)
			}
		})
	}
}
