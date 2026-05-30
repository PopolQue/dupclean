package scanner

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// TestByteScanner_Scan tests basic scanning
func TestByteScanner_Scan(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	for i := 0; i < 25; i++ {
		filename := filepath.Join(tmpDir, "file"+strconv.Itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	scanner := NewByteScanner()
	opts := Options{}

	groups, stats, err := scanner.Scan(tmpDir, opts)
	_ = groups

	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	if stats.TotalScanned != 25 {
		t.Errorf("Expected 25 files scanned, got %d", stats.TotalScanned)
	}
}
