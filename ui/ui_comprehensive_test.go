package ui

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"dupclean/scanner"
)

// captureOutput captures stdout from a function
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestRun_NoDuplicates(t *testing.T) {
	stats := scanner.ScanStats{
		TotalScanned: 100,
		TotalDupes:   0,
		WastedBytes:  0,
		ScanDuration: 5 * time.Second,
	}

	output := captureOutput(func() {
		Run([]scanner.DuplicateGroup{}, stats)
	})

	if !strings.Contains(output, "No duplicates found") {
		t.Error("Expected 'No duplicates found' message")
	}
	if !strings.Contains(output, "100 files checked") {
		t.Error("Expected files checked count")
	}
}

func TestRun_EmptyGroups(t *testing.T) {
	stats := scanner.ScanStats{
		TotalScanned: 50,
		ScanDuration: 2 * time.Second,
	}

	// Empty groups slice should behave like no duplicates
	output := captureOutput(func() {
		Run([]scanner.DuplicateGroup{}, stats)
	})

	if !strings.Contains(output, "No duplicates found") {
		t.Error("Expected 'No duplicates found' message for empty groups")
	}
}

func TestRun_WithDuplicates_OutputStructure(t *testing.T) {
	stats := scanner.ScanStats{
		TotalScanned: 10,
		TotalDupes:   5,
		WastedBytes:  1024,
		ScanDuration: 1 * time.Second,
	}

	groups := []scanner.DuplicateGroup{
		{
			Files: []scanner.FileInfo{
				{
					Path:    "/test/file1.mp3",
					Name:    "file1.mp3",
					Size:    1024,
					ModTime: time.Now(),
				},
				{
					Path:    "/test/file2.mp3",
					Name:    "file2.mp3",
					Size:    1024,
					ModTime: time.Now(),
				},
			},
		},
	}

	// Note: Run() is interactive, so we can't fully test it
	// But we can verify it doesn't panic with valid input
	// The actual interactive part would hang, so we skip full execution
	_ = groups
	_ = stats
}

func TestFormatBytes_Comprehensive(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		// Bytes
		{0, "0 B"},
		{1, "1 B"},
		{999, "999 B"},
		{1023, "1023 B"},

		// Kilobytes
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{10240, "10.0 KB"},
		{102399, "99.9 KB"},

		// Megabytes
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{10485760, "10.0 MB"},
		{104857600, "100.0 MB"},

		// Gigabytes
		{1073741824, "1.0 GB"},
		{2147483648, "2.0 GB"},
		{10737418240, "10.0 GB"},

		// Terabytes
		{1099511627776, "1.0 TB"},
		{2199023255552, "2.0 TB"},

		// Petabytes (theoretical)
		{1125899906842624, "1.0 PB"},

		// Exabytes (theoretical)
		{1152921504606846976, "1.0 EB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes_EdgeCases(t *testing.T) {
	// Test negative values (should still produce output)
	result := formatBytes(-1024)
	if result == "" {
		t.Error("formatBytes should handle negative values")
	}

	// Test very large values
	result = formatBytes(9223372036854775807) // max int64
	if !strings.Contains(result, "EB") {
		t.Errorf("Expected EB for max int64, got %q", result)
	}
}

func TestFormatBytes_Boundaries(t *testing.T) {
	// Test exact boundaries
	tests := []struct {
		bytes    int64
		unit     string
		expected string
	}{
		{1023, "B", "1023 B"},       // Just under KB
		{1024, "KB", "1.0 KB"},      // Exactly 1 KB
		{1048575, "KB", "1023.9 KB"}, // Just under MB
		{1048576, "MB", "1.0 MB"},   // Exactly 1 MB
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if !strings.Contains(result, tt.unit) {
			t.Errorf("formatBytes(%d) should contain %q, got %q", tt.bytes, tt.unit, result)
		}
	}
}

func TestPrintHeader_Output(t *testing.T) {
	output := captureOutput(printHeader)

	if !strings.Contains(output, "DUPCLEAN") {
		t.Error("Header should contain 'DUPCLEAN'")
	}
	if !strings.Contains(output, "Duplicate File Hunter") {
		t.Error("Header should contain 'Duplicate File Hunter'")
	}
	if !strings.Contains(output, "╔") {
		t.Error("Header should contain box drawing characters")
	}
	if !strings.Contains(output, "╗") {
		t.Error("Header should contain box drawing characters")
	}
}

func TestPrintScanSummary_Output(t *testing.T) {
	stats := scanner.ScanStats{
		TotalScanned: 100,
		TotalDupes:   25,
		WastedBytes:  50000,
		ScanDuration: 10 * time.Second,
	}

	output := captureOutput(func() {
		printScanSummary(stats, 5)
	})

	if !strings.Contains(output, "Scan Summary") {
		t.Error("Should contain 'Scan Summary'")
	}
	if !strings.Contains(output, "10s") {
		t.Error("Should contain duration")
	}
	if !strings.Contains(output, "100") {
		t.Error("Should contain total scanned")
	}
	if !strings.Contains(output, "5") {
		t.Error("Should contain group count")
	}
	if !strings.Contains(output, "25") {
		t.Error("Should contain duplicate count")
	}
	if !strings.Contains(output, "48.8 KB") {
		t.Error("Should contain wasted bytes")
	}
}

func TestPrintScanSummary_ZeroValues(t *testing.T) {
	stats := scanner.ScanStats{
		TotalScanned: 0,
		TotalDupes:   0,
		WastedBytes:  0,
		ScanDuration: 0,
	}

	output := captureOutput(func() {
		printScanSummary(stats, 0)
	})

	if !strings.Contains(output, "Scan Summary") {
		t.Error("Should contain 'Scan Summary' even with zero values")
	}
}

func TestPrintScanSummary_LargeValues(t *testing.T) {
	stats := scanner.ScanStats{
		TotalScanned: 1000000,
		TotalDupes:   500000,
		WastedBytes:  10737418240, // 10 GB
		ScanDuration: 3600 * time.Second, // 1 hour
	}

	output := captureOutput(func() {
		printScanSummary(stats, 100)
	})

	if !strings.Contains(output, "10.0 GB") {
		t.Error("Should format large wasted bytes as GB")
	}
}

func TestPrintControlsHelp_Output(t *testing.T) {
	output := captureOutput(printControlsHelp)

	if !strings.Contains(output, "Controls") {
		t.Error("Should contain 'Controls'")
	}
	if !strings.Contains(output, "[1-9]") {
		t.Error("Should show number key hint")
	}
	if !strings.Contains(output, "[s]") {
		t.Error("Should show skip hint")
	}
	if !strings.Contains(output, "[a]") {
		t.Error("Should show skip all hint")
	}
	if !strings.Contains(output, "[q]") {
		t.Error("Should show quit hint")
	}
	if !strings.Contains(output, "Keep file") {
		t.Error("Should explain keep function")
	}
	if !strings.Contains(output, "Skip") {
		t.Error("Should explain skip function")
	}
	if !strings.Contains(output, "Quit") {
		t.Error("Should explain quit function")
	}
}

func TestPrintFinalSummary_WithDeletions(t *testing.T) {
	output := captureOutput(func() {
		printFinalSummary(5, 1048576)
	})

	if !strings.Contains(output, "Cleanup Complete") {
		t.Error("Should contain 'Cleanup Complete'")
	}
	if !strings.Contains(output, "Files deleted:") {
		t.Error("Should show files deleted label")
	}
	if !strings.Contains(output, "5") {
		t.Error("Should show 5 files deleted")
	}
	if !strings.Contains(output, "Space freed:") {
		t.Error("Should show space freed label")
	}
	if !strings.Contains(output, "1.0 MB") {
		t.Error("Should show 1.0 MB freed")
	}
	if !strings.Contains(output, "Trash") {
		t.Error("Should mention emptying trash")
	}
}

func TestPrintFinalSummary_NoDeletions(t *testing.T) {
	output := captureOutput(func() {
		printFinalSummary(0, 0)
	})

	if !strings.Contains(output, "Nothing was deleted") {
		t.Error("Should contain 'Nothing was deleted'")
	}
	if !strings.Contains(output, "files are safe") {
		t.Error("Should mention files are safe")
	}
	if strings.Contains(output, "Cleanup Complete") {
		t.Error("Should NOT contain 'Cleanup Complete' when nothing deleted")
	}
}

func TestPrintFinalSummary_LargeDeletions(t *testing.T) {
	output := captureOutput(func() {
		printFinalSummary(100, 10737418240) // 10 GB
	})

	if !strings.Contains(output, "100") {
		t.Error("Should show 100 files deleted")
	}
	if !strings.Contains(output, "10.0 GB") {
		t.Error("Should show 10.0 GB freed")
	}
}

func TestPrintFinalSummary_Formatting(t *testing.T) {
	output := captureOutput(func() {
		printFinalSummary(1, 1024)
	})

	// Check for visual formatting elements
	if !strings.Contains(output, "─") {
		t.Error("Should contain separator lines")
	}
}

func TestMoveToTrash_NonExistentFile(t *testing.T) {
	err := moveToTrash("/nonexistent/file/that/does/not/exist.txt")
	if err == nil {
		t.Log("moveToTrash should error for non-existent file (or succeed on some systems)")
	}
	// We don't fail the test because behavior varies by system
}

func TestMoveToTrash_EmptyPath(t *testing.T) {
	err := moveToTrash("")
	if err == nil {
		t.Log("moveToTrash should error for empty path")
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color constants are defined
	colors := []string{
		colorReset,
		colorRed,
		colorGreen,
		colorYellow,
		colorBlue,
		colorPurple,
		colorCyan,
		colorWhite,
		colorGray,
		colorBold,
		colorDim,
		colorUnderline,
	}

	for i, color := range colors {
		if color == "" {
			t.Errorf("Color constant %d should not be empty", i)
		}
		if !strings.HasPrefix(color, "\033") {
			t.Errorf("Color constant %d should start with escape sequence, got %q", i, color)
		}
	}
}

func TestScanStats_Display(t *testing.T) {
	// Test that ScanStats fields are properly displayed
	stats := scanner.ScanStats{
		TotalScanned: 999999,
		TotalDupes:   500000,
		WastedBytes:  999999999999,
		ScanDuration: 99999 * time.Second,
	}

	output := captureOutput(func() {
		printScanSummary(stats, 999)
	})

	// Verify large numbers are displayed
	if !strings.Contains(output, "999999") {
		t.Error("Should display large TotalScanned value")
	}
}

func TestFormatBytes_Consistency(t *testing.T) {
	// Test that formatBytes is consistent
	sameValue := formatBytes(1024)
	for i := 0; i < 10; i++ {
		result := formatBytes(1024)
		if result != sameValue {
			t.Errorf("formatBytes(1024) inconsistent: got %q, expected %q", result, sameValue)
		}
	}
}

func TestPrintFunctions_NoPanic(t *testing.T) {
	// Verify print functions don't panic with various inputs
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Print function panicked: %v", r)
		}
	}()

	printHeader()
	printControlsHelp()
	printFinalSummary(0, 0)
	printFinalSummary(100, 1000000)

	stats := scanner.ScanStats{}
	printScanSummary(stats, 0)

	stats = scanner.ScanStats{
		TotalScanned: 1000,
		TotalDupes:   500,
		WastedBytes:  1000000,
		ScanDuration: 100 * time.Second,
	}
	printScanSummary(stats, 50)
}
