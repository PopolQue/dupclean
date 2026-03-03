package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAudioExtensions(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".wav", true},
		{".aiff", true},
		{".aif", true},
		{".mp3", true},
		{".flac", true},
		{".ogg", true},
		{".m4a", true},
		{".aac", true},
		{".opus", true},
		{".wma", true},
		{".WAV", false},
		{".MP3", false},
		{".txt", false},
		{".pdf", false},
		{".jpg", false},
		{"", false},
		{".FlAc", false},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := audioExtensions[tt.ext]
			if got != tt.expected {
				t.Errorf("audioExtensions[%q] = %v, want %v", tt.ext, got, tt.expected)
			}
		})
	}
}

func TestFileInfoStruct(t *testing.T) {
	now := time.Now()
	fi := FileInfo{
		Path:    "/path/to/file.wav",
		Name:    "file.wav",
		Size:    1024,
		ModTime: now,
		Hash:    "abc123",
	}

	if fi.Path != "/path/to/file.wav" {
		t.Errorf("Path = %q, want %q", fi.Path, "/path/to/file.wav")
	}
	if fi.Name != "file.wav" {
		t.Errorf("Name = %q, want %q", fi.Name, "file.wav")
	}
	if fi.Size != 1024 {
		t.Errorf("Size = %d, want %d", fi.Size, 1024)
	}
	if fi.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", fi.Hash, "abc123")
	}
	if !fi.ModTime.Equal(now) {
		t.Errorf("ModTime = %v, want %v", fi.ModTime, now)
	}
}

func TestDuplicateGroupStruct(t *testing.T) {
	files := []FileInfo{
		{Path: "/path/file1.wav", Name: "file1.wav", Size: 1024, Hash: "samehash"},
		{Path: "/path/file2.wav", Name: "file2.wav", Size: 1024, Hash: "samehash"},
	}

	dg := DuplicateGroup{
		Hash:  "samehash",
		Files: files,
	}

	if dg.Hash != "samehash" {
		t.Errorf("Hash = %q, want %q", dg.Hash, "samehash")
	}
	if len(dg.Files) != 2 {
		t.Errorf("len(Files) = %d, want %d", len(dg.Files), 2)
	}
}

func TestScanStatsStruct(t *testing.T) {
	stats := ScanStats{
		TotalScanned: 100,
		TotalDupes:   25,
		WastedBytes:  50000,
		ScanDuration: 10 * time.Second,
	}

	if stats.TotalScanned != 100 {
		t.Errorf("TotalScanned = %d, want %d", stats.TotalScanned, 100)
	}
	if stats.TotalDupes != 25 {
		t.Errorf("TotalDupes = %d, want %d", stats.TotalDupes, 25)
	}
	if stats.WastedBytes != 50000 {
		t.Errorf("WastedBytes = %d, want %d", stats.WastedBytes, 50000)
	}
	if stats.ScanDuration != 10*time.Second {
		t.Errorf("ScanDuration = %v, want %v", stats.ScanDuration, 10*time.Second)
	}
}

func TestFindDuplicates_EmptyFolder(t *testing.T) {
	tmpDir := t.TempDir()

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups, got %d", len(groups))
	}
	if stats.TotalScanned != 0 {
		t.Errorf("TotalScanned = %d, want 0", stats.TotalScanned)
	}
}

func TestFindDuplicates_NoAudioFiles(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file.pdf"), []byte("world"), 0644)

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups for non-audio files, got %d", len(groups))
	}
	if stats.TotalScanned != 0 {
		t.Errorf("TotalScanned = %d, want 0", stats.TotalScanned)
	}
}

func TestFindDuplicates_ScanAllFiles(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file.pdf"), []byte("world"), 0644)

	groups, stats, err := FindDuplicates(tmpDir, true, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if stats.TotalScanned != 2 {
		t.Errorf("TotalScanned = %d, want 2", stats.TotalScanned)
	}
	if len(groups) != 0 {
		t.Errorf("Expected 0 groups (no duplicates), got %d", len(groups))
	}
}

func TestFindDuplicates_NoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), []byte("content2"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.wav"), []byte("content3"), 0644)

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups (no duplicates), got %d", len(groups))
	}
	if stats.TotalScanned != 3 {
		t.Errorf("TotalScanned = %d, want 3", stats.TotalScanned)
	}
}

func TestFindDuplicates_WithDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	content1 := []byte("identical content 1")
	content2 := []byte("identical content 2")

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), content1, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), content1, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.wav"), content2, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file4.wav"), content2, 0644)

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups, got %d", len(groups))
	}
	if stats.TotalScanned != 4 {
		t.Errorf("TotalScanned = %d, want 4", stats.TotalScanned)
	}
	if stats.TotalDupes != 2 {
		t.Errorf("TotalDupes = %d, want 2", stats.TotalDupes)
	}

	for _, group := range groups {
		if len(group.Files) != 2 {
			t.Errorf("Expected 2 files per group, got %d", len(group.Files))
		}
		if group.Hash == "" {
			t.Error("Hash should not be empty")
		}
	}
}

func TestFindDuplicates_ThreeWayDuplicate(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("same content three times")

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.wav"), content, 0644)

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Files) != 3 {
		t.Errorf("Expected 3 files in group, got %d", len(groups[0].Files))
	}
	if stats.TotalDupes != 2 {
		t.Errorf("TotalDupes = %d, want 2", stats.TotalDupes)
	}
}

func TestFindDuplicates_PreservesOriginal(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("original content")

	os.WriteFile(filepath.Join(tmpDir, "original.wav"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "duplicate.wav"), content, 0644)

	groups, _, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(groups))
	}

	foundOriginal := false
	for _, f := range groups[0].Files {
		if f.Name == "original.wav" {
			foundOriginal = true
		}
		if _, err := os.Stat(f.Path); os.IsNotExist(err) {
			t.Errorf("Original file should still exist: %s", f.Path)
		}
	}
	if !foundOriginal {
		t.Error("Original file should be in the duplicate group")
	}
}

func TestFindDuplicates_SkipsHiddenFiles(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, ".hidden.wav"), []byte("hidden"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "visible.wav"), []byte("visible"), 0644)

	_, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if stats.TotalScanned != 1 {
		t.Errorf("TotalScanned = %d, want 1 (hidden files should be skipped)", stats.TotalScanned)
	}
}

func TestFindDuplicates_SkipsHiddenDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	os.Mkdir(filepath.Join(tmpDir, ".hidden"), 0755)
	os.WriteFile(filepath.Join(tmpDir, ".hidden", "file.wav"), []byte("hidden dir"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "visible.wav"), []byte("visible"), 0644)

	_, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if stats.TotalScanned != 1 {
		t.Errorf("TotalScanned = %d, want 1 (hidden dirs should be skipped)", stats.TotalScanned)
	}
}

func TestFindDuplicates_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	subDir1 := filepath.Join(tmpDir, "subdir1")
	subDir2 := filepath.Join(tmpDir, "subdir2", "nested")
	os.MkdirAll(subDir1, 0755)
	os.MkdirAll(subDir2, 0755)

	content := []byte("nested duplicate")

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), content, 0644)
	os.WriteFile(filepath.Join(subDir1, "file2.wav"), content, 0644)
	os.WriteFile(filepath.Join(subDir2, "file3.wav"), content, 0644)

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(groups))
	}
	if len(groups[0].Files) != 3 {
		t.Errorf("Expected 3 files in group, got %d", len(groups[0].Files))
	}
	if stats.TotalScanned != 3 {
		t.Errorf("TotalScanned = %d, want 3", stats.TotalScanned)
	}
}

func TestFindDuplicates_DifferentSizesNoDuplicates(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "small.wav"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "medium.wav"), []byte("ab"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "large.wav"), []byte("abc"), 0644)

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 0 {
		t.Errorf("Expected 0 groups (different sizes), got %d", len(groups))
	}
	if stats.TotalScanned != 3 {
		t.Errorf("TotalScanned = %d, want 3", stats.TotalScanned)
	}
	if stats.TotalDupes != 0 {
		t.Errorf("TotalDupes = %d, want 0", stats.TotalDupes)
	}
}

func TestFindDuplicates_WastedBytesCalculation(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("x")
	fileSize := int64(len(content))

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.wav"), content, 0644)

	_, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	expectedWasted := fileSize * 2
	if stats.WastedBytes != expectedWasted {
		t.Errorf("WastedBytes = %d, want %d (2 duplicates * file size)", stats.WastedBytes, expectedWasted)
	}
}

func TestHashFile(t *testing.T) {
	tmpDir := t.TempDir()

	testContent := []byte("test content for hashing")
	filePath := filepath.Join(tmpDir, "test.wav")
	os.WriteFile(filePath, testContent, 0644)

	hash, info, err := hashFile(filePath)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	if hash == "" {
		t.Error("hash should not be empty")
	}

	if len(hash) != 64 {
		t.Errorf("SHA-256 hash should be 64 chars, got %d", len(hash))
	}

	if info.Size() != int64(len(testContent)) {
		t.Errorf("info.Size() = %d, want %d", info.Size(), len(testContent))
	}
}

func TestHashFile_SameContentSameHash(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("identical content")

	file1 := filepath.Join(tmpDir, "file1.wav")
	file2 := filepath.Join(tmpDir, "file2.wav")
	os.WriteFile(file1, content, 0644)
	os.WriteFile(file2, content, 0644)

	hash1, _, err := hashFile(file1)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	hash2, _, err := hashFile(file2)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	if hash1 != hash2 {
		t.Errorf("Same content should produce same hash: %q != %q", hash1, hash2)
	}
}

func TestHashFile_DifferentContentDifferentHash(t *testing.T) {
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "file1.wav")
	file2 := filepath.Join(tmpDir, "file2.wav")
	os.WriteFile(file1, []byte("content 1"), 0644)
	os.WriteFile(file2, []byte("content 2"), 0644)

	hash1, _, err := hashFile(file1)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	hash2, _, err := hashFile(file2)
	if err != nil {
		t.Fatalf("hashFile() error = %v", err)
	}

	if hash1 == hash2 {
		t.Error("Different content should produce different hashes")
	}
}

func TestHashFile_NonExistentFile(t *testing.T) {
	_, _, err := hashFile("/nonexistent/path/file.wav")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestFindDuplicates_InvalidPath(t *testing.T) {
	groups, stats, err := FindDuplicates("/nonexistent/path/to/folder", false, nil, []string{}, []string{})
	if err != nil {
		t.Logf("FindDuplicates returned error for invalid path: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("Expected 0 groups for invalid path, got %d", len(groups))
	}
	if stats.TotalScanned != 0 {
		t.Errorf("Expected 0 scanned for invalid path, got %d", stats.TotalScanned)
	}
}

func TestFindDuplicates_CaseInsensitiveExtension(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("test content")

	os.WriteFile(filepath.Join(tmpDir, "file1.WAV"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.Wav"), content, 0644)

	groups, _, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group (case-insensitive extension), got %d", len(groups))
	}
	if len(groups[0].Files) != 3 {
		t.Errorf("Expected 3 files in group, got %d", len(groups[0].Files))
	}
}

func TestFindDuplicates_MixedAudioAndNonAudio(t *testing.T) {
	tmpDir := t.TempDir()

	audioContent := []byte("audio content")
	nonAudioContent := []byte("non-audio content")

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), audioContent, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), audioContent, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), nonAudioContent, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file4.txt"), nonAudioContent, 0644)

	groups, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 group (only audio files), got %d", len(groups))
	}
	if stats.TotalScanned != 2 {
		t.Errorf("TotalScanned = %d, want 2 (only audio files)", stats.TotalScanned)
	}
}

func TestFindDuplicates_ScanAllMixedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	wavContent := []byte("same wav content")
	txtContent := []byte("same txt content")

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), wavContent, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), wavContent, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file3.txt"), txtContent, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file4.txt"), txtContent, 0644)

	groups, stats, err := FindDuplicates(tmpDir, true, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("Expected 2 groups (wav and txt), got %d", len(groups))
	}
	if stats.TotalScanned != 4 {
		t.Errorf("TotalScanned = %d, want 4", stats.TotalScanned)
	}
}

func TestFindDuplicates_FileInfoPopulated(t *testing.T) {
	tmpDir := t.TempDir()

	content := []byte("test content for info")
	originalTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.Local)
	os.Chtimes(filepath.Join(tmpDir, "file1.wav"), originalTime, originalTime)
	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), content, 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.wav"), content, 0644)

	groups, _, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	for _, group := range groups {
		for _, file := range group.Files {
			if file.Name == "" {
				t.Error("FileInfo.Name should not be empty")
			}
			if file.Path == "" {
				t.Error("FileInfo.Path should not be empty")
			}
			if file.Size == 0 {
				t.Error("FileInfo.Size should not be 0")
			}
			if file.Hash == "" {
				t.Error("FileInfo.Hash should not be empty")
			}
		}
	}
}

func TestFindDuplicates_ScanDurationRecorded(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "file1.wav"), []byte("content1"), 0644)

	_, stats, err := FindDuplicates(tmpDir, false, nil, []string{}, []string{})
	if err != nil {
		t.Fatalf("FindDuplicates() error = %v", err)
	}

	if stats.ScanDuration <= 0 {
		t.Error("ScanDuration should be recorded")
	}
}
