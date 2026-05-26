package ui

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"dupclean/internal/fsutil"
	"dupclean/scanner"
)

func TestFormatBytes_UI(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{10240, "10.0 KB"},
		{102400, "100.0 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{104857600, "100.0 MB"},
		{1073741824, "1.0 GB"},
		{2147483648, "2.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := fsutil.FormatBytes(tt.bytes)
			if result != tt.expected {
				t.Errorf("fsutil.FormatBytes(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestPrintHeader_Business(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printHeader()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify header contains expected elements
	if !strings.Contains(output, "DUPCLEAN") {
		t.Error("Header should contain 'DUPCLEAN'")
	}
	if !strings.Contains(output, "Duplicate File Hunter") {
		t.Error("Header should contain 'Duplicate File Hunter'")
	}
}

func TestPrintScanSummary_Business(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stats := scanner.ScanStats{
		TotalScanned: 100,
		TotalDupes:   25,
		WastedBytes:  50000,
		ScanDuration: 10 * time.Second,
	}
	printScanSummary(stats, 5)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify summary contains expected elements
	if !strings.Contains(output, "Scan Summary") {
		t.Error("Should contain 'Scan Summary'")
	}
	if !strings.Contains(output, "100") {
		t.Error("Should contain total scanned count")
	}
	if !strings.Contains(output, "5") {
		t.Error("Should contain groups count")
	}
}

func TestPrintFinalSummary_Business_WithDeletions(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printFinalSummary(5, 1048576)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Cleanup Complete") {
		t.Error("Should contain 'Cleanup Complete'")
	}
	if !strings.Contains(output, "Files deleted:") {
		t.Error("Should show files deleted")
	}
	if !strings.Contains(output, "5") {
		t.Error("Should show 5 files deleted")
	}
	if !strings.Contains(output, "Space freed:") {
		t.Error("Should show space freed")
	}
	if !strings.Contains(output, "Trash") {
		t.Error("Should mention emptying trash")
	}
}

func TestPrintFinalSummary_Business_NoDeletions(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printFinalSummary(0, 0)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Nothing was deleted") {
		t.Error("Should contain 'Nothing was deleted'")
	}
	if !strings.Contains(output, "files are safe") {
		t.Error("Should mention files are safe")
	}
}

func TestPrintControlsHelp_Business(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printControlsHelp()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

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
}

func TestFormatBytes_EdgeCases(t *testing.T) {
	// Test zero
	if fsutil.FormatBytes(0) != "0 B" {
		t.Errorf("fsutil.FormatBytes(0) = %q, want '0 B'", fsutil.FormatBytes(0))
	}

	// Test 1 byte
	if fsutil.FormatBytes(1) != "1 B" {
		t.Errorf("fsutil.FormatBytes(1) = %q, want '1 B'", fsutil.FormatBytes(1))
	}

	// Test exact KB
	if fsutil.FormatBytes(1024) != "1.0 KB" {
		t.Errorf("fsutil.FormatBytes(1024) = %q, want '1.0 KB'", fsutil.FormatBytes(1024))
	}

	// Test exact MB
	if fsutil.FormatBytes(1048576) != "1.0 MB" {
		t.Errorf("fsutil.FormatBytes(1048576) = %q, want '1.0 MB'", fsutil.FormatBytes(1048576))
	}

	// Test exact GB
	if fsutil.FormatBytes(1073741824) != "1.0 GB" {
		t.Errorf("fsutil.FormatBytes(1073741824) = %q, want '1.0 GB'", fsutil.FormatBytes(1073741824))
	}

	// Test exact TB
	if fsutil.FormatBytes(1099511627776) != "1.0 TB" {
		t.Errorf("fsutil.FormatBytes(1099511627776) = %q, want '1.0 TB'", fsutil.FormatBytes(1099511627776))
	}
}

func TestPrintScanSummary_Business_ZeroValues(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	stats := scanner.ScanStats{
		TotalScanned: 0,
		TotalDupes:   0,
		WastedBytes:  0,
		ScanDuration: 0,
	}
	printScanSummary(stats, 0)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Scan Summary") {
		t.Error("Should contain 'Scan Summary'")
	}
}

func TestPrintFinalSummary_Business_LargeAmounts(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test with large freed amount (GB)
	printFinalSummary(100, 10737418240) // 10 GB

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "GB") {
		t.Error("Should show GB for large amounts")
	}
}
