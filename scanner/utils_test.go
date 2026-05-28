package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashFilePartial_NonExistentFile(t *testing.T) {
	_, err := hashFilePartial("/path/that/does/not/exist", 8192)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestHashFileFull_NonExistentFile(t *testing.T) {
	_, _, err := hashFileFull("/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestUtils_FilesIdentical_NonExistentFiles(t *testing.T) {
	identical, err := filesIdentical("/path/that/does/not/exist", "/another/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error when files don't exist")
	}
	if identical {
		t.Error("Non-existent files should not be considered identical")
	}
}

func TestUtils_FilesIdentical_DifferentSizes(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	_ = os.WriteFile(file1, []byte("123"), 0644)
	_ = os.WriteFile(file2, []byte("12345"), 0644)

	identical, err := filesIdentical(file1, file2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if identical {
		t.Error("Files with different sizes should not be identical")
	}
}

func TestUtils_FilesIdentical_SameSizeDifferentContent(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	_ = os.WriteFile(file1, []byte("12345"), 0644)
	_ = os.WriteFile(file2, []byte("54321"), 0644)

	identical, err := filesIdentical(file1, file2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if identical {
		t.Error("Files with different content should not be identical")
	}
}
