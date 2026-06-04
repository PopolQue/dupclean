package fsutil

import (
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"negative", -100, "n/a"},
		{"zero", 0, "0 B"},
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1024 * 1024 * 5, "5.0 MB"},
		{"gigabytes", 1024 * 1024 * 1024 * 2, "2.0 GB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024 * 3, "3.0 TB"},
		{"petabytes", 1024 * 1024 * 1024 * 1024 * 1024, "1.0 PB"},
		{"exabytes", 1024 * 1024 * 1024 * 1024 * 1024 * 1024, "1.0 EB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("FormatBytes(%d) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		{"empty", "", 0, false},
		{"days", "2d", 48 * time.Hour, false},
		{"hours", "3h", 3 * time.Hour, false},
		{"minutes", "30m", 30 * time.Minute, false},
		{"invalid", "abc", 0, true},
		{"invalidDays", "ad", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("ParseDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
