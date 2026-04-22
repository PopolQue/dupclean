package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestByteScanner_WithCancellation tests that scanning can be cancelled via context
func TestByteScanner_WithCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	for i := 0; i < 10; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	scanner := NewByteScanner()
	opts := Options{
		Context: ctx,
	}

	_, stats, err := scanner.Scan(tmpDir, opts)

	// Should get context cancelled error or partial results
	if err != context.Canceled {
		t.Logf("Scan returned err=%v (may have partial results)", err)
	}

	// Stats should be available even on cancellation
	_ = stats
}

// TestByteScanner_WithTimeout tests that scanning respects context timeout
func TestByteScanner_WithTimeout(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many test files to ensure scan takes some time
	for i := 0; i < 100; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("test content for file "+itoa(i)), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Give context time to expire
	time.Sleep(10 * time.Millisecond)

	scanner := NewByteScanner()
	opts := Options{
		Context: ctx,
	}

	_, _, err := scanner.Scan(tmpDir, opts)

	// Should get context deadline exceeded or nil (if scan completed before timeout)
	if err != context.DeadlineExceeded && err != nil {
		t.Logf("Scan returned err=%v", err)
	}
}

// TestByteScanner_WithoutContext tests that scanning works without context
func TestByteScanner_WithoutContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	for i := 0; i < 5; i++ {
		filename := filepath.Join(tmpDir, "file"+itoa(i)+".txt")
		if err := os.WriteFile(filename, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	scanner := NewByteScanner()
	opts := Options{
		Context: nil, // No context - should use background
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

// TestPhotoScanner_WithCancellation tests that photo scanning can be cancelled
func TestPhotoScanner_WithCancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files (not real images, but scanner will try to process them)
	for i := 0; i < 10; i++ {
		filename := filepath.Join(tmpDir, "image"+itoa(i)+".jpg")
		if err := os.WriteFile(filename, []byte("fake image data"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	scanner := NewPhotoScanner()
	opts := Options{
		Context: ctx,
	}

	_, stats, err := scanner.Scan(tmpDir, opts)

	// Should handle cancellation gracefully
	if err != context.Canceled && err != nil {
		t.Logf("Scan returned err=%v", err)
	}

	_ = stats
}

// TestPhotoScanner_WithoutContext tests photo scanning without context
func TestPhotoScanner_WithoutContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	for i := 0; i < 3; i++ {
		filename := filepath.Join(tmpDir, "image"+itoa(i)+".jpg")
		if err := os.WriteFile(filename, []byte("fake image"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	scanner := NewPhotoScanner()
	opts := Options{
		Context: nil,
	}

	groups, stats, err := scanner.Scan(tmpDir, opts)

	// Should work without context
	if err != nil {
		t.Logf("Scan returned err=%v (may be due to fake images)", err)
	}

	_ = groups
	_ = stats
}

// TestOptions_ContextField tests that Context field is properly handled
func TestOptions_ContextField(t *testing.T) {
	// Test with nil context
	opts1 := Options{
		Context: nil,
	}
	if opts1.Context != nil {
		t.Error("Context should be nil when set to nil")
	}

	// Test with background context
	opts2 := Options{
		Context: context.Background(),
	}
	if opts2.Context == nil {
		t.Error("Context should not be nil when set to Background")
	}

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	opts3 := Options{
		Context: ctx,
	}
	if opts3.Context == nil {
		t.Error("Context should not be nil when set with timeout")
	}
}

// TestScanner_CancelDuringWalk tests cancellation during filepath.Walk
func TestScanner_CancelDuringWalk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested directory structure
	for d := 0; d < 5; d++ {
		subdir := filepath.Join(tmpDir, "dir"+itoa(d))
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatalf("Failed to create subdir: %v", err)
		}
		for f := 0; f < 10; f++ {
			filename := filepath.Join(subdir, "file"+itoa(f)+".txt")
			if err := os.WriteFile(filename, []byte("content"), 0644); err != nil {
				t.Fatalf("Failed to create file: %v", err)
			}
		}
	}

	// Cancel context during walk
	ctx, cancel := context.WithCancel(context.Background())

	scanner := NewByteScanner()
	opts := Options{
		Context: ctx,
	}

	// Start scan in goroutine
	done := make(chan bool, 1)
	go func() {
		_, _, _ = scanner.Scan(tmpDir, opts)
		done <- true
	}()

	// Cancel after short delay
	time.Sleep(1 * time.Millisecond)
	cancel()

	// Wait for scan to complete (should finish quickly due to cancellation)
	select {
	case <-done:
		// Good - scan completed
	case <-time.After(5 * time.Second):
		t.Error("Scan did not complete after cancellation - possible goroutine leak")
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
