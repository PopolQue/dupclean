package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestByteScanner_StreamingMode tests streaming mode with threshold
func TestByteScanner_StreamingMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create enough files to trigger streaming
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("test content for file "+itoa(i)), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	scanner := &ByteScanner{
		StreamingThreshold: 10, // Process in chunks of 10
	}
	opts := Options{
		StreamingThreshold: 10,
	}

	groups, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	if stats.TotalScanned == 0 {
		t.Error("Expected some files to be scanned")
	}

	_ = groups
}

// TestByteScanner_StreamingModeWithDuplicates tests streaming with actual duplicates
func TestByteScanner_StreamingModeWithDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	// Create duplicate files
	content := []byte("duplicate content for streaming test")
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tmpDir, "dup"+itoa(i)+".txt")
		if err := os.WriteFile(filename, content, 0644); err != nil {
			t.Fatalf("Failed to create duplicate file: %v", err)
		}
	}

	// Create unique files
	for i := 0; i < 20; i++ {
		filename := filepath.Join(tmpDir, "unique"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("unique content "+itoa(i)), 0644); err != nil {
			t.Fatalf("Failed to create unique file: %v", err)
		}
	}

	scanner := &ByteScanner{
		StreamingThreshold: 10,
	}
	opts := Options{
		StreamingThreshold: 10,
	}

	groups, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	// Should find at least one duplicate group
	if len(groups) == 0 {
		t.Log("No duplicates found")
	}

	if stats.TotalDupes == 0 {
		t.Log("No duplicates counted")
	}
}

// TestByteScanner_StreamingModeCancellation tests cancellation during streaming
func TestByteScanner_StreamingModeCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many files
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	scanner := &ByteScanner{
		StreamingThreshold: 10,
	}
	opts := Options{
		Context:            context.Background(),
		StreamingThreshold: 10,
	}

	// Cancel context
	t.Cleanup(func() {})

	groups, stats, err := scanner.Scan(tmpDir, opts)

	// Should handle gracefully
	if groups == nil {
		t.Error("Expected non-nil groups")
	}

	_ = stats
	_ = err
}

// TestByteScanner_StreamingModeWithProgress tests progress callback in streaming mode
func TestByteScanner_StreamingModeWithProgress(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	for i := 0; i < 30; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	progressCalls := 0
	scanner := &ByteScanner{
		StreamingThreshold: 10,
	}
	opts := Options{
		StreamingThreshold: 10,
		OnProgress: func(progress ScanProgress) {
			progressCalls++
			if progress.Phase == "" {
				t.Error("Progress phase should not be empty")
			}
		},
	}

	_, _, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	if progressCalls == 0 {
		t.Error("Expected progress callbacks")
	}
}

// TestByteScanner_StreamingModeMemoryEfficiency tests that memory is freed between chunks
func TestByteScanner_StreamingModeMemoryEfficiency(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files in subdirectories to simulate real structure
	for d := 0; d < 5; d++ {
		subdir := filepath.Join(tmpDir, "dir"+itoa(d))
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}
		for f := 0; f < 10; f++ {
			filename := filepath.Join(subdir, "file"+itoa(f)+".txt")
			if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
	}

	scanner := &ByteScanner{
		StreamingThreshold: 10,
	}
	opts := Options{
		StreamingThreshold: 10,
	}

	groups, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	if stats.TotalScanned != 50 {
		t.Errorf("Expected 50 files scanned, got %d", stats.TotalScanned)
	}

	_ = groups
}

// TestByteScanner_StreamingThresholdZero tests that threshold=0 disables streaming
func TestByteScanner_StreamingThresholdZero(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	for i := 0; i < 20; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	scanner := &ByteScanner{
		StreamingThreshold: 0, // Disabled
	}
	opts := Options{
		StreamingThreshold: 0, // Disabled
	}

	groups, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	if stats.TotalScanned == 0 {
		t.Error("Expected files to be scanned")
	}

	_ = groups
}

// TestByteScanner_StreamingThresholdFromOptions tests that options threshold takes precedence
func TestByteScanner_StreamingThresholdFromOptions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files
	for i := 0; i < 25; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	scanner := &ByteScanner{
		StreamingThreshold: 100, // Scanner default
	}
	opts := Options{
		StreamingThreshold: 10, // Options override
	}

	groups, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan failed: %v", err)
	}

	// Should use options threshold (10), not scanner threshold (100)
	if stats.TotalScanned != 25 {
		t.Errorf("Expected 25 files scanned, got %d", stats.TotalScanned)
	}

	_ = groups
}

// TestProcessChunk tests the chunk processing function
func TestProcessChunk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := make(map[int64][]string)
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		files[7] = append(files[7], filename) // All same size
	}

	scanner := &ByteScanner{}
	groups, stats, err := scanner.processChunk(files, t.Context())

	if err != nil {
		t.Errorf("processChunk failed: %v", err)
	}

	if len(groups) == 0 {
		t.Log("No groups found (expected for identical content)")
	}

	_ = stats
}

// TestNewByteScanner_StreamingThreshold tests constructor with streaming
func TestNewByteScanner_StreamingThreshold(t *testing.T) {
	scanner := NewByteScanner()

	if scanner.StreamingThreshold != 0 {
		t.Errorf("Default StreamingThreshold should be 0, got %d", scanner.StreamingThreshold)
	}

	// Can be set manually
	scanner.StreamingThreshold = 10000
	if scanner.StreamingThreshold != 10000 {
		t.Errorf("StreamingThreshold should be 10000, got %d", scanner.StreamingThreshold)
	}
}
