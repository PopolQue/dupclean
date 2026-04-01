package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPhotoScanner_Scan_EmptyDirectory(t *testing.T) {
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

func TestPhotoScanner_Scan_NonImageFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-image files
	txt := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(txt, []byte("not an image"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	scanner := NewPhotoScanner()
	opts := Options{
		IncludeHidden: false,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	// Photo scanner may or may not count non-image files
	// The important thing is it doesn't crash
	_ = result
	_ = stats
}

func TestPhotoScanner_Scan_IgnoreExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	png1 := filepath.Join(tmpDir, "image1.png")
	jpg := filepath.Join(tmpDir, "image.jpg")

	if err := os.WriteFile(png1, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create PNG: %v", err)
	}
	if err := os.WriteFile(jpg, []byte("fake jpg"), 0644); err != nil {
		t.Fatalf("Failed to create JPG: %v", err)
	}

	scanner := NewPhotoScanner()
	opts := Options{
		IncludeHidden:    false,
		IgnoreExtensions: []string{".jpg"},
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	// Should only scan PNG
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (.jpg ignored), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestPhotoScanner_Scan_MinSize(t *testing.T) {
	tmpDir := t.TempDir()

	// Create small and large files
	small := filepath.Join(tmpDir, "small.png")
	large := filepath.Join(tmpDir, "large.png")

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := os.WriteFile(small, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create small PNG: %v", err)
	}
	if err := os.WriteFile(large, append(pngHeader, make([]byte, 1000)...), 0644); err != nil {
		t.Fatalf("Failed to create large PNG: %v", err)
	}

	scanner := NewPhotoScanner()
	opts := Options{
		IncludeHidden: false,
		MinSize:       100,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	// Should only scan large file
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (small filtered), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestPhotoScanner_Scan_HiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	visible := filepath.Join(tmpDir, "visible.png")
	hidden := filepath.Join(tmpDir, ".hidden.png")

	if err := os.WriteFile(visible, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create visible PNG: %v", err)
	}
	if err := os.WriteFile(hidden, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create hidden PNG: %v", err)
	}

	scanner := NewPhotoScanner()
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

func TestPhotoScanner_Scan_IncludeHidden(t *testing.T) {
	tmpDir := t.TempDir()

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	visible := filepath.Join(tmpDir, "visible.png")
	hidden := filepath.Join(tmpDir, ".hidden.png")

	if err := os.WriteFile(visible, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create visible PNG: %v", err)
	}
	if err := os.WriteFile(hidden, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create hidden PNG: %v", err)
	}

	scanner := NewPhotoScanner()
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

func TestPhotoScanner_Scan_IgnoreFolders(t *testing.T) {
	tmpDir := t.TempDir()

	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	rootFile := filepath.Join(tmpDir, "root.png")
	ignoredDir := filepath.Join(tmpDir, "@eaDir")
	ignoredFile := filepath.Join(ignoredDir, "syno.png")

	if err := os.WriteFile(rootFile, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create root PNG: %v", err)
	}
	if err := os.MkdirAll(ignoredDir, 0755); err != nil {
		t.Fatalf("Failed to create ignored dir: %v", err)
	}
	if err := os.WriteFile(ignoredFile, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create ignored PNG: %v", err)
	}

	scanner := NewPhotoScanner()
	opts := Options{
		IncludeHidden: false,
		IgnoreFolders: []string{ignoredDir},
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (@eaDir ignored), got %d", stats.TotalScanned)
	}
	_ = result
}

func TestPhotoScanner_Scan_SimilarityThreshold(t *testing.T) {
	tmpDir := t.TempDir()

	// Create identical PNG files (will have same hash)
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	file1 := filepath.Join(tmpDir, "copy1.png")
	file2 := filepath.Join(tmpDir, "copy2.png")

	if err := os.WriteFile(file1, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	scanner := NewPhotoScanner()
	scanner.SimilarityPct = 90 // 90% similarity threshold
	opts := Options{
		IncludeHidden: false,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 2 {
		t.Errorf("Expected 2 files scanned, got %d", stats.TotalScanned)
	}
	// May or may not find duplicates depending on hash
	_ = result
}

func TestPhotoScanner_Scan_LowSimilarity(t *testing.T) {
	tmpDir := t.TempDir()

	// Create identical files
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	file1 := filepath.Join(tmpDir, "copy1.png")
	file2 := filepath.Join(tmpDir, "copy2.png")

	if err := os.WriteFile(file1, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(file2, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	scanner := NewPhotoScanner()
	scanner.SimilarityPct = 50 // Low threshold - should find more matches
	opts := Options{
		IncludeHidden: false,
	}

	result, stats, err := scanner.Scan(tmpDir, opts)

	if err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if stats.TotalScanned != 2 {
		t.Errorf("Expected 2 files scanned, got %d", stats.TotalScanned)
	}
	_ = result
}

func TestNewPhotoScanner_DefaultSimilarity(t *testing.T) {
	scanner := NewPhotoScanner()
	if scanner.SimilarityPct != 90 {
		t.Errorf("Expected default SimilarityPct 90, got %d", scanner.SimilarityPct)
	}
}

func TestPhotoScanner_OptionsOverride(t *testing.T) {
	tmpDir := t.TempDir()
	scanner := NewPhotoScanner()

	opts := Options{
		SimilarityPct: 95, // Override default
	}

	// The scanner should use the option's similarity
	_, _, _ = scanner.Scan(tmpDir, opts)

	if scanner.SimilarityPct != 95 {
		t.Errorf("Expected SimilarityPct 95 after options override, got %d", scanner.SimilarityPct)
	}
}
