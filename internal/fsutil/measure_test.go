package fsutil

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Test MeasureResult structure
func TestMeasureResult_Struct(t *testing.T) {
	result := &MeasureResult{
		TotalSize: 1024,
		FileCount: 10,
		DirCount:  5,
		Entries:   []EntryInfo{},
	}

	if result.TotalSize != 1024 {
		t.Errorf("TotalSize = %d, want 1024", result.TotalSize)
	}
	if result.FileCount != 10 {
		t.Errorf("FileCount = %d, want 10", result.FileCount)
	}
	if result.DirCount != 5 {
		t.Errorf("DirCount = %d, want 5", result.DirCount)
	}
	if result.Entries == nil {
		t.Error("Entries should not be nil")
	}
}

// Test EntryInfo structure
func TestEntryInfo_Struct(t *testing.T) {
	now := time.Now()
	info := EntryInfo{
		Path:    "/test/path/file.txt",
		Size:    1024,
		ModTime: now,
		IsDir:   false,
	}

	if info.Path != "/test/path/file.txt" {
		t.Errorf("Path = %q, want %q", info.Path, "/test/path/file.txt")
	}
	if info.Size != 1024 {
		t.Errorf("Size = %d, want 1024", info.Size)
	}
	if info.ModTime != now {
		t.Error("ModTime mismatch")
	}
	if info.IsDir {
		t.Error("IsDir should be false")
	}
}

// Test EntryInfo for directory
func TestEntryInfo_Directory(t *testing.T) {
	info := EntryInfo{
		Path:    "/test/path/dir",
		Size:    4096,
		ModTime: time.Now(),
		IsDir:   true,
	}

	if !info.IsDir {
		t.Error("IsDir should be true")
	}
}

// Test MeasureFile with existing file
func TestMeasureFile_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("Hello, World!")

	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, err := MeasureFile(testFile)
	if err != nil {
		t.Fatalf("MeasureFile failed: %v", err)
	}

	if info.Path != testFile {
		t.Errorf("Path = %q, want %q", info.Path, testFile)
	}
	if info.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", info.Size, len(content))
	}
	if info.IsDir {
		t.Error("IsDir should be false for a file")
	}
}

// Test MeasureFile with non-existent file
func TestMeasureFile_NonExistent(t *testing.T) {
	_, err := MeasureFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("MeasureFile should return an error for non-existent file")
	}
}

// Test MeasureFile with directory
func TestMeasureFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()

	info, err := MeasureFile(tmpDir)
	if err != nil {
		t.Fatalf("MeasureFile failed: %v", err)
	}

	if !info.IsDir {
		t.Error("IsDir should be true for a directory")
	}
	if info.Path != tmpDir {
		t.Errorf("Path = %q, want %q", info.Path, tmpDir)
	}
}

// Test MeasureDir with empty directory
func TestMeasureDir_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	result, err := MeasureDir(tmpDir, nil, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result == nil {
		t.Fatal("MeasureDir should return a non-nil result")
	}
	if result.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0 for empty directory", result.TotalSize)
	}
	if result.FileCount != 0 {
		t.Errorf("FileCount = %d, want 0", result.FileCount)
	}
	if result.DirCount != 0 {
		t.Errorf("DirCount = %d, want 0", result.DirCount)
	}
}

// Test MeasureDir with files
func TestMeasureDir_WithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := map[string]int64{
		"file1.txt": 100,
		"file2.txt": 200,
		"file3.txt": 300,
	}

	for name, size := range files {
		content := make([]byte, size)
		for i := range content {
			content[i] = byte(i % 256)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	result, err := MeasureDir(tmpDir, nil, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	expectedTotal := int64(0)
	for _, size := range files {
		expectedTotal += size
	}

	if result.FileCount != 3 {
		t.Errorf("FileCount = %d, want 3", result.FileCount)
	}
	if result.TotalSize != expectedTotal {
		t.Errorf("TotalSize = %d, want %d", result.TotalSize, expectedTotal)
	}
}

// Test MeasureDir with subdirectories
func TestMeasureDir_WithSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectories
	subdirs := []string{"sub1", "sub2", "sub1/nested"}
	for _, subdir := range subdirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, subdir), 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subdir, err)
		}
	}

	// Create files in subdirectories
	files := map[string]int64{
		"file1.txt":             100,
		"sub1/file2.txt":        200,
		"sub2/file3.txt":        300,
		"sub1/nested/file4.txt": 400,
	}

	for name, size := range files {
		content := make([]byte, size)
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	result, err := MeasureDir(tmpDir, nil, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result.FileCount != 4 {
		t.Errorf("FileCount = %d, want 4", result.FileCount)
	}
	// DirCount should count subdirectories (not root)
	if result.DirCount < 3 {
		t.Errorf("DirCount = %d, want at least 3", result.DirCount)
	}
}

// Test MeasureDir with pattern matching
func TestMeasureDir_WithPatternMatching(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with different extensions
	files := map[string]int64{
		"file1.txt": 100,
		"file2.log": 200,
		"file3.txt": 300,
		"file4.tmp": 400,
		"file5.txt": 500,
	}

	for name, size := range files {
		content := make([]byte, size)
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	// Match only .txt files
	result, err := MeasureDir(tmpDir, []string{"*.txt"}, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result.FileCount != 3 {
		t.Errorf("FileCount = %d, want 3 (only .txt files)", result.FileCount)
	}

	expectedSize := files["file1.txt"] + files["file3.txt"] + files["file5.txt"]
	if result.TotalSize != expectedSize {
		t.Errorf("TotalSize = %d, want %d", result.TotalSize, expectedSize)
	}
}

// Test MeasureDir with multiple patterns
func TestMeasureDir_MultiplePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]int64{
		"file1.txt": 100,
		"file2.log": 200,
		"file3.tmp": 300,
		"file4.dat": 400,
	}

	for name, size := range files {
		content := make([]byte, size)
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to create %s: %v", name, err)
		}
	}

	// Match .txt and .log files
	result, err := MeasureDir(tmpDir, []string{"*.txt", "*.log"}, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result.FileCount != 2 {
		t.Errorf("FileCount = %d, want 2", result.FileCount)
	}

	expectedSize := files["file1.txt"] + files["file2.log"]
	if result.TotalSize != expectedSize {
		t.Errorf("TotalSize = %d, want %d", result.TotalSize, expectedSize)
	}
}

// Test MeasureDir with minAge filter
func TestMeasureDir_WithMinAge(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tmpDir, "old.txt")
	content := []byte("old content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Modify the file's timestamp to be old
	oldTime := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(testFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set file time: %v", err)
	}

	// Measure with minAge of 1 hour - should include the old file
	result, err := MeasureDir(tmpDir, nil, time.Hour)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1 (old file)", result.FileCount)
	}

	// Create a new file
	newFile := filepath.Join(tmpDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	// Measure with minAge of 1 hour - should still only include the old file
	result2, err := MeasureDir(tmpDir, nil, time.Hour)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result2.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1 (only old file, new file is too recent)", result2.FileCount)
	}
}

// Test MeasureDir with no minAge filter
func TestMeasureDir_NoMinAge(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	result, err := MeasureDir(tmpDir, nil, 0) // minAge = 0 means no filter
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result.FileCount != 5 {
		t.Errorf("FileCount = %d, want 5", result.FileCount)
	}
}

// Test MeasureDir with inaccessible directory
func TestMeasureDir_InaccessibleDirectory(t *testing.T) {
	// Try to measure a directory that likely doesn't exist or is inaccessible
	_, err := MeasureDir("/root/nonexistent/directory", nil, 0)
	// Should not panic, may return error or empty result
	if err != nil {
		t.Logf("MeasureDir returned error for inaccessible dir: %v", err)
	}
}

// Test computeDirSize helper function
func TestComputeDirSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	totalSize := int64(0)
	for i := 0; i < 5; i++ {
		size := int64(100 * (i + 1))
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		content := make([]byte, size)
		if err := os.WriteFile(filename, content, 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		totalSize += size
	}

	size, err := computeDirSize(tmpDir)
	if err != nil {
		t.Fatalf("computeDirSize failed: %v", err)
	}

	// Note: Directory metadata size may vary by filesystem
	// We just verify the size is at least the sum of files
	if size < totalSize {
		t.Errorf("Size = %d, want at least %d", size, totalSize)
	}
}

// Test computeDirSize with empty directory
func TestComputeDirSize_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	size, err := computeDirSize(tmpDir)
	if err != nil {
		t.Fatalf("computeDirSize failed: %v", err)
	}

	// Note: Empty directories may have some metadata size (e.g., 64 bytes on APFS)
	// We just verify it's a small value
	if size < 0 {
		t.Errorf("Size should be non-negative, got %d", size)
	}
}

// Test computeDirSize with nested directories
func TestComputeDirSize_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure
	subdir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create files
	files := map[string]int64{
		"file1.txt":        100,
		"subdir/file2.txt": 200,
	}

	totalSize := int64(0)
	for name, size := range files {
		content := make([]byte, size)
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		totalSize += size
	}

	size, err := computeDirSize(tmpDir)
	if err != nil {
		t.Fatalf("computeDirSize failed: %v", err)
	}

	// Note: Directory metadata size may vary by filesystem
	// We just verify the size is at least the sum of files
	if size < totalSize {
		t.Errorf("Size = %d, want at least %d", size, totalSize)
	}
}

// Test MeasureResult with entries
func TestMeasureResult_Entries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	for i := 0; i < 3; i++ {
		filename := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		content := make([]byte, 100)
		if err := os.WriteFile(filename, content, 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	result, err := MeasureDir(tmpDir, nil, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if len(result.Entries) == 0 {
		t.Error("Entries should not be empty")
	}

	// Verify entry structure
	for _, entry := range result.Entries {
		if entry.Path == "" {
			t.Error("Entry path should not be empty")
		}
		if entry.Size < 0 {
			t.Errorf("Entry size should be non-negative, got %d", entry.Size)
		}
		if entry.ModTime.IsZero() {
			t.Error("Entry ModTime should not be zero")
		}
	}
}

// Test pattern matching edge cases
func TestMeasureDir_PatternEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with various names
	files := []string{
		"test.txt",
		"TEST.TXT",
		"test.log",
		".hidden",
		"noextension",
	}

	for _, name := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Pattern matching is case-sensitive on most systems
	result, err := MeasureDir(tmpDir, []string{"*.txt"}, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	// Should match test.txt (case depends on OS)
	t.Logf("Pattern *.txt matched %d files", result.FileCount)
}

// Test MeasureDir with wildcard patterns
func TestMeasureDir_WildcardPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	files := map[string]int64{
		"test1.txt": 100,
		"test2.txt": 200,
		"temp1.log": 300,
		"temp2.log": 400,
		"other.dat": 500,
	}

	for name, size := range files {
		content := make([]byte, size)
		if err := os.WriteFile(filepath.Join(tmpDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Test test* pattern
	result, err := MeasureDir(tmpDir, []string{"test*"}, 0)
	if err != nil {
		t.Fatalf("MeasureDir failed: %v", err)
	}

	if result.FileCount != 2 {
		t.Errorf("FileCount = %d, want 2 for test* pattern", result.FileCount)
	}
}
