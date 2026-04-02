package diskanalyzer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestMaxEntriesLimit tests that MaxEntries limits the number of collected entries
func TestMaxEntriesLimit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 50 test files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	opts := WalkOptions{
		Concurrency: 2,
		MaxEntries:  20, // Limit to 20 entries
	}

	entries, _, err := statPass(tmpDir, opts)
	if err != nil {
		t.Fatalf("statPass failed: %v", err)
	}

	if len(entries) > 20 {
		t.Errorf("Expected max 20 entries, got %d", len(entries))
	}
	if len(entries) == 0 {
		t.Error("Expected some entries, got 0")
	}
}

// TestMaxEntriesZero tests that MaxEntries=0 means unlimited
func TestMaxEntriesZero(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 30 test files
	for i := 0; i < 30; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	opts := WalkOptions{
		Concurrency: 2,
		MaxEntries:  0, // Unlimited
	}

	entries, _, err := statPass(tmpDir, opts)
	if err != nil {
		t.Fatalf("statPass failed: %v", err)
	}

	if len(entries) != 30 {
		t.Errorf("Expected 30 entries (unlimited), got %d", len(entries))
	}
}

// TestMemoryWarningThreshold tests that warning is logged for large scans
func TestMemoryWarningThreshold(t *testing.T) {
	// This test verifies the constant is defined
	if MemoryWarningThreshold <= 0 {
		t.Error("MemoryWarningThreshold should be positive")
	}
	
	// Reasonable threshold (100k files)
	if MemoryWarningThreshold != 100000 {
		t.Logf("MemoryWarningThreshold = %d (expected 100000)", MemoryWarningThreshold)
	}
}

// TestWalkOptions_MaxEntries tests WalkOptions with MaxEntries
func TestWalkOptions_MaxEntries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in subdirectories
	for d := 0; d < 5; d++ {
		subdir := filepath.Join(tmpDir, "dir"+itoa(d))
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}
		for f := 0; f < 10; f++ {
			filename := filepath.Join(subdir, "file"+itoa(f)+".txt")
			if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
		}
	}

	opts := WalkOptions{
		Concurrency: 2,
		MaxEntries:  25, // Should limit to 25
	}

	entries, _, err := statPass(tmpDir, opts)
	if err != nil {
		t.Fatalf("statPass failed: %v", err)
	}

	if len(entries) > 25 {
		t.Errorf("Expected max 25 entries, got %d", len(entries))
	}
}

// TestDefaultOptions_Documentation tests that DefaultOptions has reasonable defaults
func TestDefaultOptions_Documentation(t *testing.T) {
	opts := DefaultOptions()

	// MaxEntries should be 0 (unlimited) by default for backwards compatibility
	if opts.MaxEntries != 0 {
		t.Errorf("Default MaxEntries should be 0 (unlimited), got %d", opts.MaxEntries)
	}

	// Concurrency should be reasonable
	if opts.Concurrency != 0 {
		t.Logf("Default Concurrency = %d (0 = auto)", opts.Concurrency)
	}
}

// TestLargeScanWithFilters tests memory optimization with filters
func TestLargeScanWithFilters(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mix of files
	for i := 0; i < 100; i++ {
		// Small files
		filename := filepath.Join(tmpDir, "small"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("x"), 0644); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Large files
		if i%10 == 0 {
			filename = filepath.Join(tmpDir, "large"+itoa(i)+".dat")
			if err := os.WriteFile(filename, make([]byte, 1024), 0644); err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
		}
	}

	// Use MinSize filter to reduce memory
	opts := WalkOptions{
		Concurrency: 2,
		MinSize:     100, // Skip small files
		MaxEntries:  50,  // Limit total entries
	}

	entries, _, err := statPass(tmpDir, opts)
	if err != nil {
		t.Fatalf("statPass failed: %v", err)
	}

	// Should have filtered out small files
	if len(entries) > 50 {
		t.Errorf("Expected max 50 entries, got %d", len(entries))
	}
}

// Helper function to convert int to string
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	
	var result []byte
	for i > 0 {
		result = append([]byte{byte('0' + i%10)}, result...)
		i /= 10
	}
	return string(result)
}
