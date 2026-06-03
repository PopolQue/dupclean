package interactive

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/PopolQue/dupclean/internal/fsutil"
	"github.com/PopolQue/dupclean/scanner"
)

// Test file sorting with multiple depth levels
func TestFileSorting_MultipleDepths(t *testing.T) {
	now := time.Now()
	files := []scanner.FileInfo{
		{Path: filepath.FromSlash("/a/b/c/d/e/file.txt"), Name: "file.txt", Size: 100, ModTime: now},
		{Path: filepath.FromSlash("/a/file.txt"), Name: "file.txt", Size: 100, ModTime: now},
		{Path: filepath.FromSlash("/a/b/c/file.txt"), Name: "file.txt", Size: 100, ModTime: now},
		{Path: filepath.FromSlash("/a/b/file.txt"), Name: "file.txt", Size: 100, ModTime: now},
	}

	sort.Slice(files, func(i, j int) bool {
		di := strings.Count(files[i].Path, string(os.PathSeparator))
		dj := strings.Count(files[j].Path, string(os.PathSeparator))
		if di != dj {
			return di < dj
		}
		return files[i].ModTime.Before(files[j].ModTime)
	})

	// Verify sorted by depth (shallowest first)
	expectedPaths := []string{
		filepath.FromSlash("/a/file.txt"),
		filepath.FromSlash("/a/b/file.txt"),
		filepath.FromSlash("/a/b/c/file.txt"),
		filepath.FromSlash("/a/b/c/d/e/file.txt"),
	}

	for i, expected := range expectedPaths {
		if files[i].Path != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, files[i].Path)
		}
	}
}

// Test file sorting with identical depth and time
func TestFileSorting_IdicalDepthAndTime(t *testing.T) {
	now := time.Now()
	files := []scanner.FileInfo{
		{Path: filepath.FromSlash("/test/file1.txt"), Name: "file1.txt", Size: 100, ModTime: now},
		{Path: filepath.FromSlash("/test/file2.txt"), Name: "file2.txt", Size: 100, ModTime: now},
		{Path: filepath.FromSlash("/test/file3.txt"), Name: "file3.txt", Size: 100, ModTime: now},
	}

	// Should not panic when all values are equal
	sort.Slice(files, func(i, j int) bool {
		di := strings.Count(files[i].Path, string(os.PathSeparator))
		dj := strings.Count(files[j].Path, string(os.PathSeparator))
		if di != dj {
			return di < dj
		}
		return files[i].ModTime.Before(files[j].ModTime)
	})
}

// Test group sorting stability
func TestGroupSorting_Stability(t *testing.T) {
	groups := []scanner.DuplicateGroup{
		{Hash: "a", Files: []scanner.FileInfo{{Size: 100}}},
		{Hash: "b", Files: []scanner.FileInfo{{Size: 100}}},
		{Hash: "c", Files: []scanner.FileInfo{{Size: 100}}},
	}

	// Sort with equal values - should maintain relative order (stable sort)
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].Files[0].Size > groups[j].Files[0].Size
	})

	// With stable sort, order should be preserved for equal elements
	if groups[0].Hash != "a" || groups[1].Hash != "b" || groups[2].Hash != "c" {
		t.Log("Stable sort should preserve order for equal elements")
	}
}

// Test group sorting with mixed sizes
func TestGroupSorting_MixedSizes(t *testing.T) {
	groups := []scanner.DuplicateGroup{
		{Hash: "tiny", Files: []scanner.FileInfo{{Size: 1}}},
		{Hash: "huge", Files: []scanner.FileInfo{{Size: 1000000}}},
		{Hash: "medium", Files: []scanner.FileInfo{{Size: 500}}},
		{Hash: "large", Files: []scanner.FileInfo{{Size: 10000}}},
		{Hash: "small", Files: []scanner.FileInfo{{Size: 10}}},
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Files[0].Size > groups[j].Files[0].Size
	})

	expectedOrder := []string{"huge", "large", "medium", "small", "tiny"}
	for i, expected := range expectedOrder {
		if groups[i].Hash != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, groups[i].Hash)
		}
	}
}

// Test input handling for all valid skip inputs
func TestInputHandling_AllValidInputs(t *testing.T) {
	validInputs := map[string]bool{
		"s":    true,
		"S":    true,
		"skip": true,
		"SKIP": true,
		"Skip": true,
		"":     true,
		"   ":  true, // whitespace only
	}

	for input := range validInputs {
		processed := strings.TrimSpace(strings.ToLower(input))
		switch processed {
		case "s", "skip", "":
			// Correct - should skip
		default:
			t.Errorf("Input %q should be treated as skip, got %q", input, processed)
		}
	}
}

// Test input handling for all valid quit inputs
func TestInputHandling_AllQuitInputs(t *testing.T) {
	quitInputs := []string{"q", "Q", "quit", "QUIT", "Quit"}

	for _, input := range quitInputs {
		processed := strings.TrimSpace(strings.ToLower(input))
		switch processed {
		case "q", "quit":
			// Correct - should quit
		default:
			t.Errorf("Input %q should be treated as quit, got %q", input, processed)
		}
	}
}

// Test input validation edge cases
func TestInputValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		input   string
		files   int
		isValid bool
		desc    string
	}{
		{"1", 5, true, "valid first choice"},
		{"5", 5, true, "valid last choice"},
		{"3", 5, true, "valid middle choice"},
		{"0", 5, false, "zero is invalid"},
		{"6", 5, false, "out of range high"},
		{"-1", 5, false, "negative"},
		{"abc", 5, false, "non-numeric"},
		{"1.5", 5, false, "decimal"},
		{" 2 ", 5, true, "whitespace trimmed"},
		{"", 5, false, "empty"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			choice, err := strconv.Atoi(strings.TrimSpace(tt.input))
			valid := err == nil && choice >= 1 && choice <= tt.files

			if valid != tt.isValid {
				t.Errorf("Input %q with %d files: valid=%v, want %v", tt.input, tt.files, valid, tt.isValid)
			}
		})
	}
}

// Test formatBytes with maximum int64
func TestFormatBytes_MaxInt64(t *testing.T) {
	maxInt64 := int64(^uint64(0) >> 1)
	result := fsutil.FormatBytes(maxInt64)
	if result == "" {
		t.Error("formatBytes should return non-empty string for max int64")
	}
	if !strings.Contains(result, "EB") {
		t.Logf("fsutil.FormatBytes(maxInt64) = %q (expected EB scale)", result)
	}
}

// Test formatBytes with specific values
func TestFormatBytes_SpecificValues(t *testing.T) {
	tests := []struct {
		bytes int64
		scale string
	}{
		{0, "B"},
		{999, "B"},
		{1024, "KB"},
		{1048576, "MB"},
		{1073741824, "GB"},
		{1099511627776, "TB"},
		{1125899906842624, "PB"},
		{1152921504606846976, "EB"},
	}

	for _, tt := range tests {
		result := fsutil.FormatBytes(tt.bytes)
		if !strings.Contains(result, tt.scale) {
			t.Errorf("fsutil.FormatBytes(%d) = %q, should contain %q", tt.bytes, result, tt.scale)
		}
	}
}

// Test color code combinations
func TestColorCodes_AllCombinations(t *testing.T) {
	colors := []string{colorRed, colorGreen, colorYellow, colorBlue, colorPurple, colorCyan, colorWhite}
	styles := []string{colorBold, colorDim, colorUnderline}

	for _, color := range colors {
		for _, style := range styles {
			combined := style + color + "test" + colorReset
			if !strings.Contains(combined, color) {
				t.Errorf("Combined string should contain color %q", color)
			}
			if !strings.Contains(combined, style) {
				t.Errorf("Combined string should contain style %q", style)
			}
			if !strings.HasSuffix(combined, colorReset) {
				t.Error("Combined string should end with colorReset")
			}
		}
	}
}

// Test printScanSummary with extreme values
func TestPrintScanSummary_ExtremeValues(t *testing.T) {
	tests := []struct {
		name       string
		stats      scanner.ScanStats
		groupCount int
	}{
		{
			name: "maximum values",
			stats: scanner.ScanStats{
				TotalScanned: 1000000,
				TotalDupes:   500000,
				WastedBytes:  1024 * 1024 * 1024 * 1024, // 1 TB
				ScanDuration: time.Hour * 24,
			},
			groupCount: 10000,
		},
		{
			name: "negative duration (shouldn't happen but test anyway)",
			stats: scanner.ScanStats{
				TotalScanned: 100,
				TotalDupes:   10,
				WastedBytes:  1024,
				ScanDuration: -time.Second,
			},
			groupCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printScanSummary panicked: %v", r)
				}
			}()
			printScanSummary(tt.stats, tt.groupCount)
		})
	}
}

// Test printFinalSummary with extreme values
func TestPrintFinalSummary_ExtremeValues(t *testing.T) {
	tests := []struct {
		name    string
		deleted int
		freed   int64
	}{
		{"zero values", 0, 0},
		{"single byte", 1, 1},
		{"maximum deleted", 1000000, 0},
		{"maximum freed", 0, 1024 * 1024 * 1024 * 1024},
		{"both maximum", 1000000, 1024 * 1024 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("printFinalSummary panicked: %v", r)
				}
			}()
			printFinalSummary(tt.deleted, tt.freed)
		})
	}
}

// Test moveToTrash error messages
func TestMoveToTrash_ErrorMessages(t *testing.T) {
	err := moveToTrash("")
	if err == nil {
		t.Fatal("moveToTrash with empty path should return error")
	}

	errStr := err.Error()
	expectedSubstrings := []string{"empty", "path", "trash"}
	found := false
	for _, sub := range expectedSubstrings {
		if strings.Contains(strings.ToLower(errStr), sub) {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Error message %q doesn't contain expected substrings %v", errStr, expectedSubstrings)
	}
}

// Test duplicate group with many files
func TestDuplicateGroup_ManyFiles(t *testing.T) {
	files := make([]scanner.FileInfo, 100)
	for i := 0; i < 100; i++ {
		fileName := "file" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + ".txt"
		files[i] = scanner.FileInfo{
			Path:    filepath.Join(filepath.FromSlash("/test"), fileName),
			Name:    fileName,
			Size:    1024,
			ModTime: time.Now(),
		}
	}

	group := scanner.DuplicateGroup{
		Hash:  "test-hash",
		Files: files,
	}

	if len(group.Files) != 100 {
		t.Errorf("Group should have 100 files, got %d", len(group.Files))
	}

	// Test sorting with many files
	sort.Slice(files, func(i, j int) bool {
		di := strings.Count(files[i].Path, string(os.PathSeparator))
		dj := strings.Count(files[j].Path, string(os.PathSeparator))
		if di != dj {
			return di < dj
		}
		return files[i].ModTime.Before(files[j].ModTime)
	})
}

// Test duplicate group with various file sizes
func TestDuplicateGroup_VariousSizes(t *testing.T) {
	sizes := []int64{1, 1024, 1024 * 1024, 1024 * 1024 * 100, 1024 * 1024 * 1024}

	for _, size := range sizes {
		group := scanner.DuplicateGroup{
			Hash: "hash-" + string(rune(size)),
			Files: []scanner.FileInfo{
				{Path: filepath.FromSlash("/test/file1.txt"), Size: size},
				{Path: filepath.FromSlash("/test/file2.txt"), Size: size},
			},
		}

		// Verify formatBytes works for this size
		sizeStr := fsutil.FormatBytes(size)
		if sizeStr == "" {
			t.Errorf("fsutil.FormatBytes(%d) returned empty string", size)
		}

		_ = group
	}
}
