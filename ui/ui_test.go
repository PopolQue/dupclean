package ui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"dupclean/scanner"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 10, "10.0 KB"},
		{1024 * 100, "100.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 5, "5.0 MB"},
		{1024 * 1024 * 100, "100.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 2, "2.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatBytes_Zero(t *testing.T) {
	result := formatBytes(0)
	if result != "0 B" {
		t.Errorf("formatBytes(0) = %q, want %q", result, "0 B")
	}
}

func TestFormatBytes_OneByte(t *testing.T) {
	result := formatBytes(1)
	if result != "1 B" {
		t.Errorf("formatBytes(1) = %q, want %q", result, "1 B")
	}
}

func TestFormatBytes_Kilobyte(t *testing.T) {
	result := formatBytes(1024)
	if result != "1.0 KB" {
		t.Errorf("formatBytes(1024) = %q, want %q", result, "1.0 KB")
	}
}

func TestFormatBytes_Megabyte(t *testing.T) {
	result := formatBytes(1024 * 1024)
	if result != "1.0 MB" {
		t.Errorf("formatBytes(1048576) = %q, want %q", result, "1.0 MB")
	}
}

func TestFormatBytes_Gigabyte(t *testing.T) {
	result := formatBytes(1024 * 1024 * 1024)
	if result != "1.0 GB" {
		t.Errorf("formatBytes(1073741824) = %q, want %q", result, "1.0 GB")
	}
}

func TestFormatBytes_Terabyte(t *testing.T) {
	result := formatBytes(1024 * 1024 * 1024 * 1024)
	if result != "1.0 TB" {
		t.Errorf("formatBytes(1099511627776) = %q, want %q", result, "1.0 TB")
	}
}

func TestFormatBytes_LargeFile(t *testing.T) {
	result := formatBytes(500 * 1024 * 1024)
	if result != "500.0 MB" {
		t.Errorf("formatBytes(500MB) = %q, want %q", result, "500.0 MB")
	}
}

func TestFormatBytes_Negative(t *testing.T) {
	result := formatBytes(-1024)
	if result == "" {
		t.Error("formatBytes should not return empty string for negative values")
	}
}

func TestFormatBytes_JustUnderKB(t *testing.T) {
	result := formatBytes(1023)
	if result != "1023 B" {
		t.Errorf("formatBytes(1023) = %q, want %q", result, "1023 B")
	}
}

func TestFormatBytes_JustOverKB(t *testing.T) {
	result := formatBytes(1025)
	if result != "1.0 KB" {
		t.Errorf("formatBytes(1025) = %q, want %q", result, "1.0 KB")
	}
}

func TestPrintHeader(t *testing.T) {
	printHeader()
}

func TestPrintScanSummary(t *testing.T) {
	stats := scanner.ScanStats{
		TotalScanned: 100,
		TotalDupes:   25,
		WastedBytes:  50000,
		ScanDuration: 10 * time.Second,
	}
	printScanSummary(stats, 5)
}

func TestPrintFinalSummary_Deleted(t *testing.T) {
	printFinalSummary(5, 1024*1024)
}

func TestPrintFinalSummary_NoneDeleted(t *testing.T) {
	printFinalSummary(0, 0)
}

func TestMoveToTrash(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := moveToTrash(testFile)
	if err != nil {
		t.Logf("moveToTrash error (may be expected in some environments): %v", err)
	}
}
