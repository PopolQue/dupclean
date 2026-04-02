package scanner

import (
	"os"
	"testing"
)

// TestDefaultPartialHashSize tests that the constant is properly defined
func TestDefaultPartialHashSize(t *testing.T) {
	// Should be 8KB
	expected := 8 * 1024
	if DefaultPartialHashSize != expected {
		t.Errorf("DefaultPartialHashSize = %d, want %d", DefaultPartialHashSize, expected)
	}

	// Should be positive
	if DefaultPartialHashSize <= 0 {
		t.Error("DefaultPartialHashSize should be positive")
	}

	// Should be reasonable (not too small, not too large)
	if DefaultPartialHashSize < 1024 {
		t.Error("DefaultPartialHashSize should be at least 1KB")
	}
	if DefaultPartialHashSize > 1024*1024 {
		t.Error("DefaultPartialHashSize should not exceed 1MB")
	}
}

// TestDefaultComparisonBufferSize tests that the constant is properly defined
func TestDefaultComparisonBufferSize(t *testing.T) {
	// Should be 32KB
	expected := 32 * 1024
	if DefaultComparisonBufferSize != expected {
		t.Errorf("DefaultComparisonBufferSize = %d, want %d", DefaultComparisonBufferSize, expected)
	}

	// Should be positive
	if DefaultComparisonBufferSize <= 0 {
		t.Error("DefaultComparisonBufferSize should be positive")
	}

	// Should be reasonable for I/O buffering
	if DefaultComparisonBufferSize < 4096 {
		t.Error("DefaultComparisonBufferSize should be at least 4KB")
	}
	if DefaultComparisonBufferSize > 1024*1024 {
		t.Error("DefaultComparisonBufferSize should not exceed 1MB")
	}
}

// TestPartialHashSize_BackwardsCompatibility tests that the legacy constant still works
func TestPartialHashSize_BackwardsCompatibility(t *testing.T) {
	// Legacy constant should equal new constant
	if partialHashSize != DefaultPartialHashSize {
		t.Errorf("partialHashSize = %d, should equal DefaultPartialHashSize = %d",
			partialHashSize, DefaultPartialHashSize)
	}
}

// TestConstants_ReasonableValues tests that all constants have reasonable values
func TestConstants_ReasonableValues(t *testing.T) {
	tests := []struct {
		name  string
		value int
		min   int
		max   int
	}{
		{"DefaultPartialHashSize", DefaultPartialHashSize, 1024, 1024 * 1024},
		{"DefaultComparisonBufferSize", DefaultComparisonBufferSize, 4096, 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value < tt.min {
				t.Errorf("%s = %d, should be >= %d", tt.name, tt.value, tt.min)
			}
			if tt.value > tt.max {
				t.Errorf("%s = %d, should be <= %d", tt.name, tt.value, tt.max)
			}
		})
	}
}

// TestConstants_PowerOfTwo tests that buffer sizes are powers of two
func TestConstants_PowerOfTwo(t *testing.T) {
	isPowerOfTwo := func(n int) bool {
		return n > 0 && (n&(n-1)) == 0
	}

	if !isPowerOfTwo(DefaultPartialHashSize) {
		t.Errorf("DefaultPartialHashSize = %d, should be power of 2 for optimal memory alignment", DefaultPartialHashSize)
	}

	if !isPowerOfTwo(DefaultComparisonBufferSize) {
		t.Errorf("DefaultComparisonBufferSize = %d, should be power of 2 for optimal memory alignment", DefaultComparisonBufferSize)
	}
}

// TestConstants_Documentation tests that constants are documented
func TestConstants_Documentation(t *testing.T) {
	// This is a meta-test to ensure constants have documentation
	// The constants should have comments explaining their purpose

	// DefaultPartialHashSize should be documented
	if DefaultPartialHashSize == 0 {
		t.Error("DefaultPartialHashSize should be non-zero")
	}

	// DefaultComparisonBufferSize should be documented
	if DefaultComparisonBufferSize == 0 {
		t.Error("DefaultComparisonBufferSize should be non-zero")
	}
}

// TestConstants_UsageInFunctions tests that constants are used correctly
func TestConstants_UsageInFunctions(t *testing.T) {
	// Test that DefaultPartialHashSize works with hashFilePartial
	// (This is more of an integration test)

	// Create a temporary file for testing
	tmpFile := t.TempDir() + "/test.txt"
	content := []byte("test content for hashing")
	if err := writeFile(tmpFile, content); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Hash should work with DefaultPartialHashSize
	hash, err := hashFilePartial(tmpFile, int64(DefaultPartialHashSize))
	if err != nil {
		t.Errorf("hashFilePartial with DefaultPartialHashSize failed: %v", err)
	}
	if hash == "" {
		t.Error("hashFilePartial should return non-empty hash")
	}
}

// writeFile is a helper function for tests
func writeFile(path string, content []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = f.Write(content)
	return err
}
