package cleaner

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDeleteEntry_Exhaustive(t *testing.T) {
	oldRemoveAll := osRemoveAll
	oldTrash := moveToTrash
	oldAbs := absPath
	defer func() {
		osRemoveAll = oldRemoveAll
		moveToTrash = oldTrash
		absPath = oldAbs
	}()

	// Mock absPath to return path as-is
	absPath = func(path string) (string, error) { return path, nil }

	entry := EntryInfo{Path: "test.txt", Size: 100}

	t.Run("PermanentSuccess", func(t *testing.T) {
		osRemoveAll = func(path string) error { return nil }
		deleted, freed, skipped, err := deleteEntry(entry, true)
		if err != nil || deleted != 1 || freed != 100 || skipped {
			t.Errorf("Unexpected result: %d, %d, %v, %v", deleted, freed, skipped, err)
		}
	})

	t.Run("PermanentInUse", func(t *testing.T) {
		osRemoveAll = func(path string) error { return errors.New("file is busy") }
		deleted, freed, skipped, err := deleteEntry(entry, true)
		if err != nil || deleted != 0 || freed != 0 || !skipped {
			t.Errorf("Unexpected result: %d, %d, %v, %v", deleted, freed, skipped, err)
		}
	})

	t.Run("PermanentError", func(t *testing.T) {
		osRemoveAll = func(path string) error { return errors.New("fatal disk error") }
		deleted, freed, skipped, err := deleteEntry(entry, true)
		if err == nil || deleted != 0 || freed != 0 || skipped {
			t.Errorf("Unexpected result: %d, %d, %v, %v", deleted, freed, skipped, err)
		}
	})

	t.Run("TrashInUse", func(t *testing.T) {
		moveToTrash = func(path string) error { return errors.New("file in use") }
		deleted, freed, skipped, err := deleteEntry(entry, false)
		if err != nil || deleted != 0 || freed != 0 || !skipped {
			t.Errorf("Unexpected result: %d, %d, %v, %v", deleted, freed, skipped, err)
		}
	})
	
	t.Run("EmptyPath", func(t *testing.T) {
		deleted, _, _, err := deleteEntry(EntryInfo{Path: ""}, true)
		if err == nil || deleted != 0 {
			t.Errorf("Expected error for empty path")
		}
	})
}

func TestDeleteEntry_Permanent(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
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
}

func TestDeleteEntry_Trash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Create test file
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete to trash
	// Mock success to avoid CI issues
	oldTrash := moveToTrash
	moveToTrash = func(path string) error { return nil }
	defer func() { moveToTrash = oldTrash }()

	deleted, freed, skipped, err := deleteEntry(EntryInfo{
		Path: testFile,
		Size: 12,
	}, false)

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
}
