package cleaner

import (
	"strings"
	"testing"
)

func TestGetTotalSize(t *testing.T) {
	tests := []struct {
		name     string
		targets  []*CleanTarget
		expected int64
	}{
		{
			name:     "empty targets",
			targets:  []*CleanTarget{},
			expected: 0,
		},
		{
			name: "single target",
			targets: []*CleanTarget{
				{TotalSize: 1024},
			},
			expected: 1024,
		},
		{
			name: "multiple targets",
			targets: []*CleanTarget{
				{TotalSize: 1024},
				{TotalSize: 2048},
				{TotalSize: 512},
			},
			expected: 3584,
		},
		{
			name:     "nil targets",
			targets:  nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTotalSize(tt.targets)
			if result != tt.expected {
				t.Errorf("getTotalSize() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 512, "512 B"},
		{"1 KB", 1024, "1.00 KB"},
		{"1.5 KB", 1536, "1.50 KB"},
		{"1 MB", 1024 * 1024, "1.00 MB"},
		{"1.5 MB", int64(1024 * 1024 * 1.5), "1.50 MB"},
		{"1 GB", 1024 * 1024 * 1024, "1.00 GB"},
		{"2.5 GB", int64(1024 * 1024 * 1024 * 2.5), "2.50 GB"},
		{"1 TB", 1024 * 1024 * 1024 * 1024, "1.00 TB"},
		{"large TB", 1024 * 1024 * 1024 * 1024 * 2, "2.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatSize_Precision(t *testing.T) {
	result := formatSize(1100)
	if !strings.Contains(result, "KB") {
		t.Errorf("formatSize(1100) should contain 'KB', got %q", result)
	}

	result = formatSize(1100 * 1024)
	if !strings.Contains(result, "MB") {
		t.Errorf("formatSize(1126400) should contain 'MB', got %q", result)
	}
}

func TestGetSelectedTargets(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", Selected: true, TotalSize: 100},
			{ID: "target2", Selected: false, TotalSize: 200},
			{ID: "target3", Selected: true, TotalSize: 300},
			{ID: "target4", Selected: true, TotalSize: 0}, // Should be skipped
		},
	}

	selected := getSelectedTargets(result)

	if len(selected) != 2 {
		t.Errorf("Expected 2 selected targets, got %d", len(selected))
	}

	for _, s := range selected {
		if !s.Selected {
			t.Error("Selected target should have Selected=true")
		}
		if s.TotalSize == 0 {
			t.Error("Selected target should have TotalSize > 0")
		}
	}
}

func TestGetSelectedTargets_None(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", Selected: false},
			{ID: "target2", Selected: false},
		},
	}

	selected := getSelectedTargets(result)

	if len(selected) != 0 {
		t.Errorf("Expected 0 selected targets, got %d", len(selected))
	}
}

func TestGetSelectedTargets_Empty(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{},
	}

	selected := getSelectedTargets(result)

	if len(selected) != 0 {
		t.Errorf("Expected 0 selected targets for empty list, got %d", len(selected))
	}
}
