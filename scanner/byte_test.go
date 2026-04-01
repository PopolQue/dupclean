package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestByteScanner_Scan_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	scanner := NewByteScanner()

	opts := Options{
		IncludeHidden: false,
		MinSize:       0,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(result))
	}
	if stats.TotalScanned != 0 {
		t.Errorf("Expected 0 files scanned, got %d", stats.TotalScanned)
	}
}

func TestByteScanner_Scan_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "single.txt")

	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	scanner := NewByteScanner()
	opts := Options{
		IncludeHidden: false,
		MinSize:       0,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 groups (no duplicates), got %d", len(result))
	}
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned, got %d", stats.TotalScanned)
	}
}

func TestByteScanner_Scan_DuplicateFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create duplicate files
	content := []byte("duplicate content")
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")

	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	scanner := NewByteScanner()
	opts := Options{
		IncludeHidden: false,
		MinSize:       0,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if len(result) == 0 {
		t.Error("Expected at least 1 duplicate group")
	}
	if stats.TotalScanned != 2 {
		t.Errorf("Expected 2 files scanned, got %d", stats.TotalScanned)
	}
}

func TestByteScanner_Scan_MinSizeFilter(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files of different sizes
	small := filepath.Join(tmpDir, "small.txt")
	large := filepath.Join(tmpDir, "large.txt")

	if err := os.WriteFile(small, []byte("sm"), 0644); err != nil {
		t.Fatalf("Failed to create small file: %v", err)
	}
	if err := os.WriteFile(large, []byte("larger content here"), 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	scanner := NewByteScanner()
	opts := Options{
		IncludeHidden: false,
		MinSize:       10, // Only files >= 10 bytes
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (large only), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestByteScanner_Scan_IgnoreExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	txt := filepath.Join(tmpDir, "file.txt")
	log := filepath.Join(tmpDir, "file.log")

	if err := os.WriteFile(txt, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}
	if err := os.WriteFile(log, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create log file: %v", err)
	}

	scanner := NewByteScanner()
	opts := Options{
		IncludeHidden:    false,
		IgnoreExtensions: []string{".log"},
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (.log ignored), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestByteScanner_Scan_HiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	visible := filepath.Join(tmpDir, "visible.txt")
	hidden := filepath.Join(tmpDir, ".hidden.txt")

	if err := os.WriteFile(visible, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create visible file: %v", err)
	}
	if err := os.WriteFile(hidden, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	scanner := NewByteScanner()
	opts := Options{
		IncludeHidden: false,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (hidden excluded), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestByteScanner_Scan_IncludeHidden(t *testing.T) {
	tmpDir := t.TempDir()

	visible := filepath.Join(tmpDir, "visible.txt")
	hidden := filepath.Join(tmpDir, ".hidden.txt")

	if err := os.WriteFile(visible, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create visible file: %v", err)
	}
	if err := os.WriteFile(hidden, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create hidden file: %v", err)
	}

	scanner := NewByteScanner()
	opts := Options{
		IncludeHidden: true,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 2 {
		t.Errorf("Expected 2 files scanned (hidden included), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestByteScanner_Scan_IgnoreFolders(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in root and in ignored folder
	rootFile := filepath.Join(tmpDir, "root.txt")
	ignoredDir := filepath.Join(tmpDir, "node_modules")
	ignoredFile := filepath.Join(ignoredDir, "package.txt")

	if err := os.WriteFile(rootFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create root file: %v", err)
	}
	if err := os.MkdirAll(ignoredDir, 0755); err != nil {
		t.Fatalf("Failed to create ignored dir: %v", err)
	}
	if err := os.WriteFile(ignoredFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create ignored file: %v", err)
	}

	scanner := NewByteScanner()
	opts := Options{
		IncludeHidden: false,
		IgnoreFolders: []string{ignoredDir},
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (node_modules ignored), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestByteScanner_Scan_NonExistentPath(t *testing.T) {
	scanner := NewByteScanner()
	opts := Options{}

	// filepath.Walk returns nil for non-existent paths (no files to walk)
	result, stats, err := scanner.Scan("/nonexistent/path", opts)

	if err != nil {
		t.Logf("Scan() error (may be expected): %v", err)
	}
	if len(result) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(result))
	}
	if stats.TotalScanned != 0 {
		t.Errorf("Expected 0 files scanned, got %d", stats.TotalScanned)
	}
}

func TestByteScanner_Scan_UnreadableDirectory(t *testing.T) {
	// This test may not work in all environments
	// Skip if running as root (root can read anything)
	if os.Geteuid() == 0 {
		t.Skip("Skipping unreadable directory test when running as root")
	}

	tmpDir := t.TempDir()
	unreadableDir := filepath.Join(tmpDir, "unreadable")

	if err := os.MkdirAll(unreadableDir, 0000); err != nil {
		t.Fatalf("Failed to create unreadable dir: %v", err)
	}
	defer os.Chmod(unreadableDir, 0755) // Restore permissions for cleanup

	scanner := NewByteScanner()
	opts := Options{}

	// Should not panic, should skip unreadable directories
	result, _, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Logf("Scan() error (may be expected): %v", err)
	}
	_ = result
}

func TestNewByteScanner(t *testing.T) {
	scanner := NewByteScanner()
	if scanner == nil {
		t.Fatal("NewByteScanner() returned nil")
	}
}
