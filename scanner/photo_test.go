package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPhotoScanner(t *testing.T) {
	scanner := NewPhotoScanner()
	if scanner == nil {
		t.Fatal("NewPhotoScanner() returned nil")
	}
	if scanner.SimilarityPct != 90 {
		t.Errorf("Expected default SimilarityPct 90, got %d", scanner.SimilarityPct)
	}
}

func TestPhotoScanner_Scan_NoFiles(t *testing.T) {
	tmpDir := t.TempDir()
	scanner := NewPhotoScanner()

	opts := Options{
		IncludeHidden: false,
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

func TestPhotoScanner_Scan_WithImageFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test image files (simple valid PNG headers)
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	writeFile(t, tmpDir, "image1.png", string(pngHeader))
	writeFile(t, tmpDir, "image2.png", string(pngHeader))

	scanner := NewPhotoScanner()
	opts := Options{
		IncludeHidden: false,
	}
	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned == 0 {
		t.Error("Expected files to be scanned")
	}
	// May or may not find duplicates depending on hash similarity
	t.Logf("Found %d duplicate groups", len(result))
}

func TestPhotoScanner_Scan_IgnoreExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	writeFile(t, tmpDir, "image1.png", string(pngHeader))
	writeFile(t, tmpDir, "ignore.jpg", string(pngHeader))

	scanner := NewPhotoScanner()
	opts := Options{
		IncludeHidden:     false,
		IgnoreExtensions:  []string{".jpg"},
	}
	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	// Should only scan PNG, not JPG
	if stats.TotalScanned > 1 {
		t.Errorf("Expected 1 file scanned (JPG ignored), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestComputePerceptualHash_FileNotFound(t *testing.T) {
	_, _, err := computePerceptualHash("/nonexistent/file.png")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestComputePerceptualHash_InvalidImage(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "invalid.png")

	// Write invalid image data
	if err := os.WriteFile(testFile, []byte("not an image"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, _, err := computePerceptualHash(testFile)
	if err == nil {
		t.Error("Expected error for invalid image")
	}
}

func TestGetScanner(t *testing.T) {
	scanner, found := GetScanner("photo")
	if !found {
		t.Error("Expected photo scanner to be registered")
	}
	if scanner == nil {
		t.Error("Expected non-nil photo scanner")
	}

	scanner, found = GetScanner("audio")
	if !found {
		t.Error("Expected audio scanner to be registered")
	}
	if scanner == nil {
		t.Error("Expected non-nil audio scanner")
	}

	scanner, found = GetScanner("unknown")
	if found {
		t.Error("Expected false for unknown scanner mode")
	}
	if scanner != nil {
		t.Error("Expected nil for unknown scanner mode")
	}
}

func TestAvailableModes(t *testing.T) {
	modes := AvailableModes()
	if len(modes) == 0 {
		t.Error("Expected at least one available mode")
	}

	// Should include photo and audio modes
	foundPhoto := false
	foundAudio := false
	for _, mode := range modes {
		if mode == "photo" {
			foundPhoto = true
		}
		if mode == "audio" {
			foundAudio = true
		}
	}
	if !foundPhoto {
		t.Error("Expected 'photo' mode to be available")
	}
	if !foundAudio {
		t.Error("Expected 'audio' mode to be available")
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}
