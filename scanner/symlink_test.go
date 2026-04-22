package scanner

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestSymlinkDetection_ByteScanner tests that symlinks are skipped
func TestSymlinkDetection_ByteScanner(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("real content"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create a symlink
	symlink := filepath.Join(tmpDir, "link.txt")
	if err := os.Symlink(realFile, symlink); err != nil {
		// Skip test on Windows (symlinks require admin privileges)
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}
		t.Fatalf("Failed to create symlink: %v", err)
	}

	scanner := NewByteScanner()
	groups, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find the real file, not the symlink
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned, got %d", stats.TotalScanned)
	}

	// Verify no duplicate groups (symlink shouldn't be counted)
	if len(groups) != 0 {
		t.Errorf("Expected 0 duplicate groups, got %d", len(groups))
	}
}

// TestSymlinkDetection_PhotoScanner tests that symlinks are skipped in photo scanner
func TestSymlinkDetection_PhotoScanner(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a real image file (minimal PNG)
	realFile := filepath.Join(tmpDir, "real.png")
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if err := os.WriteFile(realFile, pngHeader, 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create a symlink
	symlink := filepath.Join(tmpDir, "link.png")
	if err := os.Symlink(realFile, symlink); err != nil {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}
		t.Fatalf("Failed to create symlink: %v", err)
	}

	scanner := NewPhotoScanner()
	groups, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find the real file, not the symlink
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned, got %d", stats.TotalScanned)
	}

	// Verify no duplicate groups (symlink shouldn't be counted)
	if len(groups) != 0 {
		t.Errorf("Expected 0 duplicate groups, got %d", len(groups))
	}
}

// TestHardLinkDetection_ByteScanner tests that hard links are skipped
func TestHardLinkDetection_ByteScanner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping hard link test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("real content"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create a hard link
	hardLink := filepath.Join(tmpDir, "hardlink.txt")
	if err := os.Link(realFile, hardLink); err != nil {
		t.Fatalf("Failed to create hard link: %v", err)
	}

	scanner := NewByteScanner()
	groups, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find one file (hard link should be skipped)
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (hard link skipped), got %d", stats.TotalScanned)
	}

	// Verify no duplicate groups
	if len(groups) != 0 {
		t.Errorf("Expected 0 duplicate groups, got %d", len(groups))
	}
}

// TestSymlinkInSubdirectory_ByteScanner tests symlinks in subdirectories
func TestSymlinkInSubdirectory_ByteScanner(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create real files in different locations
	realFile1 := filepath.Join(tmpDir, "file1.txt")
	realFile2 := filepath.Join(subDir, "file2.txt")
	if err := os.WriteFile(realFile1, []byte("content 1"), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	if err := os.WriteFile(realFile2, []byte("content 2"), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Create symlink in subdir pointing to file1
	symlink := filepath.Join(subDir, "link.txt")
	if err := os.Symlink(realFile1, symlink); err != nil {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping symlink test on Windows")
		}
		t.Fatalf("Failed to create symlink: %v", err)
	}

	scanner := NewByteScanner()
	_, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should find 2 real files, not the symlink
	if stats.TotalScanned != 2 {
		t.Errorf("Expected 2 files scanned, got %d", stats.TotalScanned)
	}
}

// TestMaliciousSymlink_ByteScanner tests that symlinks to sensitive files are skipped
func TestMaliciousSymlink_ByteScanner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a real file in tmpDir
	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("real content"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create a symlink pointing outside the scan directory
	// This simulates a malicious symlink attack
	maliciousLink := filepath.Join(tmpDir, "malicious.txt")
	if err := os.Symlink("/etc/passwd", maliciousLink); err != nil {
		t.Skipf("Could not create symlink: %v", err)
	}

	scanner := NewByteScanner()
	_, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find the real file, not follow the malicious symlink
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned (malicious symlink skipped), got %d", stats.TotalScanned)
	}
}

// TestMultipleSymlinks_ByteScanner tests multiple symlinks to same file
func TestMultipleSymlinks_ByteScanner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a real file
	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("real content"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create multiple symlinks to the same file
	for i := 0; i < 5; i++ {
		symlink := filepath.Join(tmpDir, "link"+string(rune('0'+i))+".txt")
		if err := os.Symlink(realFile, symlink); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}
	}

	scanner := NewByteScanner()
	_, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find the real file, not the symlinks
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned, got %d", stats.TotalScanned)
	}
}

// TestBrokenSymlink_ByteScanner tests that broken symlinks are handled gracefully
func TestBrokenSymlink_ByteScanner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a broken symlink (target doesn't exist)
	brokenLink := filepath.Join(tmpDir, "broken.txt")
	if err := os.Symlink("/nonexistent/file", brokenLink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Create a real file
	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("real content"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	scanner := NewByteScanner()
	_, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find the real file, broken symlink should be skipped
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned, got %d", stats.TotalScanned)
	}
}

// TestSymlinkLoop_ByteScanner tests that symlink loops don't cause infinite recursion
func TestSymlinkLoop_ByteScanner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tmpDir := t.TempDir()

	// Create a symlink loop: link1 -> link2 -> link1
	link1 := filepath.Join(tmpDir, "link1.txt")
	link2 := filepath.Join(tmpDir, "link2.txt")

	// Create link1 pointing to link2 (which doesn't exist yet)
	if err := os.Symlink(link2, link1); err != nil {
		t.Fatalf("Failed to create link1: %v", err)
	}

	// Create link2 pointing to link1
	if err := os.Symlink(link1, link2); err != nil {
		t.Fatalf("Failed to create link2: %v", err)
	}

	// Create a real file
	realFile := filepath.Join(tmpDir, "real.txt")
	if err := os.WriteFile(realFile, []byte("real content"), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// This should complete without hanging
	scanner := NewByteScanner()
	_, stats, err := scanner.Scan(tmpDir, Options{})
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// Should only find the real file
	if stats.TotalScanned != 1 {
		t.Errorf("Expected 1 file scanned, got %d", stats.TotalScanned)
	}
}

// TestGetInode tests the getInode helper function
func TestGetInode(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	inode, ok := getInode(info)

	// On Unix systems, should return valid inode
	// On Windows, should return (0, false)
	if runtime.GOOS != "windows" {
		if !ok {
			t.Error("getInode should return true on Unix systems")
		}
		if inode == 0 {
			t.Error("getInode should return non-zero inode on Unix systems")
		}
	} else {
		if ok {
			t.Error("getInode should return false on Windows")
		}
		if inode != 0 {
			t.Error("getInode should return 0 on Windows")
		}
	}
}
