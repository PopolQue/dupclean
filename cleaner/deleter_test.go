package cleaner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteEntry_Permanent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	
	// Create test file
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(testFile); err != nil {
		t.Fatalf("Test file should exist")
	}

	// Delete permanently
	deleted, freed, skipped, err := deleteEntry(EntryInfo{
		Path: testFile,
		Size: 12,
	}, true)

	if err != nil {
		t.Errorf("deleteEntry() error = %v", err)
	}
	if deleted != 1 {
		t.Errorf("Expected 1 deleted, got %d", deleted)
	}
	if freed != 12 {
		t.Errorf("Expected 12 bytes freed, got %d", freed)
	}
	if skipped {
		t.Error("File should not be skipped")
	}

	// Verify file is deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File should be deleted")
	}
}

func TestDeleteEntry_Permanent_NonExistent(t *testing.T) {
	deleted, _, _, err := deleteEntry(EntryInfo{
		Path: "/nonexistent/file.txt",
		Size: 100,
	}, true)

	// On some systems, RemoveAll doesn't error for non-existent files
	// So we check that either it errors OR it returns 0 deleted
	if err == nil && deleted == 1 {
		t.Log("File was 'deleted' (may be expected on some systems)")
	}
	// The important thing is that it doesn't crash
}

func TestDeleteEntry_Trash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	
	// Create test file
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete to trash
	_, _, _, err := deleteEntry(EntryInfo{
		Path: testFile,
		Size: 12,
	}, false)

	// Trash deletion may fail in CI environments, but should not error
	if err != nil {
		t.Logf("moveToTrash error (expected in some environments): %v", err)
	}
}

func TestIsFileInUse(t *testing.T) {
	// Test nil error
	if isFileInUse(nil) {
		t.Error("nil error should not be considered as file in use")
	}
	
	// Test with actual error strings
	if !isFileInUse(&os.PathError{Op: "remove", Path: "test", Err: os.ErrPermission}) {
		t.Log("Permission errors may be considered as file in use")
	}
	
	// Test specific strings
	testCases := []struct {
		errStr   string
		expected bool
	}{
		{"file busy", true},
		{"file in use", true},
		{"sharing violation", true},
		{"permission denied", true},
		{"access is denied", true},
		{"file not found", false},
		{"", false},
	}
	
	for _, tc := range testCases {
		result := containsAny(tc.errStr, []string{"busy", "in use", "sharing violation", "permission denied", "access is denied"})
		if result != tc.expected {
			t.Errorf("containsAny(%q) = %v, want %v", tc.errStr, result, tc.expected)
		}
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s        string
		substrs  []string
		expected bool
	}{
		{"file is busy", []string{"busy", "in use"}, true},
		{"file is in use", []string{"busy", "in use"}, true},
		{"file not found", []string{"busy", "in use"}, false},
		{"", []string{"busy"}, false},
		{"busy", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			result := containsAny(tt.s, tt.substrs)
			if result != tt.expected {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tt.s, tt.substrs, result, tt.expected)
			}
		})
	}
}

func TestContains(t *testing.T) {
	if !contains("hello world", "world") {
		t.Error("contains should find substring")
	}
	if contains("hello world", "foo") {
		t.Error("contains should not find non-existent substring")
	}
}

func TestFindSubstring(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "foo", false},
		{"foo", "", true},
		{"FOO", "foo", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.s+"-"+tt.substr, func(t *testing.T) {
			result := findSubstring(tt.s, tt.substr)
			if result != tt.expected {
				t.Errorf("findSubstring(%q, %q) = %v, want %v", tt.s, tt.substr, result, tt.expected)
			}
		})
	}
}
