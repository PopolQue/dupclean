package gui

import (
	"errors"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/PopolQue/dupclean/scanner"
)

func TestVerifyDeletionSafety_Errors(t *testing.T) {
	mockAbsErr := func(path string) (string, error) { return "", errors.New("abs error") }
	mockHomeErr := func() (string, error) { return "", errors.New("home error") }
	mockAbs := func(path string) (string, error) { return path, nil }

	// Test abs error
	err := verifyDeletionSafety("/path", mockAbsErr, mockHomeErr)
	if err == nil {
		t.Error("Expected error from resolver")
	}

	// Test home error (should still allow non-home path)
	err = verifyDeletionSafety("/path", mockAbs, mockHomeErr)
	if err != nil {
		t.Errorf("Expected no error when home lookup fails, got %v", err)
	}
}

func TestVerifyDeletionSafety(t *testing.T) {
	mockAbs := func(path string) (string, error) { return path, nil }
	mockHome := func() (string, error) { return "/home/user", nil }

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"Empty path", "", true},
		{"Root directory", "/", true},
		{"Home directory", "/home/user", true},
		{"Protected path", "/etc", true},
		{"Safe path", "/home/user/documents/file.txt", false},
		{"Invalid path", "/invalid", false},
	}

	protectedPaths := []string{
		"/bin", "/boot", "/dev", "/home", "/lib", "/lib64",
		"/media", "/mnt", "/opt", "/proc", "/root", "/run", "/sbin",
		"/srv", "/sys", "/tmp", "/usr", "/var",
		"C:\\Windows", "C:\\Program Files", "C:\\Program Files (x86)", "C:\\Users",
	}

	for _, p := range protectedPaths {
		tests = append(tests, struct {
			name    string
			path    string
			wantErr bool
		}{"Protected path " + p, p, true})
	}

	tests = append(tests, struct {
		name    string
		path    string
		wantErr bool
	}{"Windows Root", `C:\`, true})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := verifyDeletionSafety(tt.path, mockAbs, mockHome)
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyDeletionSafety() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPrepareScanResults(t *testing.T) {
	groups := []scanner.DuplicateGroup{
		{
			Hash: "abc",
			Files: []scanner.FileInfo{
				{Path: "/a/b/file1", Name: "file1", Size: 100, ModTime: time.Now()},
				{Path: "/a/file2", Name: "file2", Size: 100, ModTime: time.Now().Add(-time.Hour)},
			},
		},
	}

	sortedGroups, selections := prepareScanResults(groups)

	if len(sortedGroups[0].Files) != 2 {
		t.Fatal("Expected 2 files in group")
	}

	if sortedGroups[0].Files[0].Path != "/a/file2" {
		t.Errorf("Expected first file to be /a/file2, got %s", sortedGroups[0].Files[0].Path)
	}

	if !selections[0][0] {
		t.Error("Expected first file to be selected")
	}
}

func TestUpdateScanButtonState(t *testing.T) {
	_ = test.NewApp()
	state := &AppState{}
	scanBtn := widget.NewButton("", nil)
	scanBtn.Disable()

	// Test enabling
	updateScanButtonState(state, "/path/to/folder", scanBtn)
	if state.FolderPath != "/path/to/folder" {
		t.Errorf("FolderPath = %s, want /path/to/folder", state.FolderPath)
	}
	if scanBtn.Disabled() {
		t.Error("Button should be enabled when text is not empty")
	}

	// Test disabling
	updateScanButtonState(state, "", scanBtn)
	if state.FolderPath != "" {
		t.Errorf("FolderPath = %s, want empty", state.FolderPath)
	}
	if !scanBtn.Disabled() {
		t.Error("Button should be disabled when text is empty")
	}
}
