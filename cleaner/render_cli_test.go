package cleaner

import (
	"testing"
)

func TestRenderCLI(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{
				ID:          "test-target",
				Category:    "Test",
				Label:       "Test Target",
				Description: "A test target",
				Paths:       []string{"/test"},
				Patterns:    []string{"*"},
				Risk:        RiskSafe,
				TotalSize:   1024,
				Selected:    false,
			},
		},
		TotalSize: 1024,
	}

	opts := CLIOptions{}

	// RenderCLI writes to stdout, just verify it doesn't panic
	RenderCLI(result, opts)
}

func TestRenderCLI_EmptyResult(t *testing.T) {
	result := &ScanResult{
		Targets:   []*CleanTarget{},
		TotalSize: 0,
	}

	opts := CLIOptions{}

	// Should not panic
	RenderCLI(result, opts)
}

func TestRenderCLI_WithModerateRisk(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{
				ID:          "moderate-target",
				Category:    "Test",
				Label:       "Moderate Risk",
				Description: "A moderate risk target",
				Paths:       []string{"/test"},
				Patterns:    []string{"*"},
				Risk:        RiskModerate,
				TotalSize:   2048,
				Selected:    false,
			},
		},
		TotalSize: 2048,
	}

	opts := CLIOptions{}

	// Should not panic
	RenderCLI(result, opts)
}

func TestRenderCLI_MultipleTargets(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{
				ID:        "target1",
				Category:  "Browser",
				Label:     "Chrome Cache",
				Paths:     []string{"/test/chrome"},
				Patterns:  []string{"*"},
				Risk:      RiskSafe,
				TotalSize: 1024,
			},
			{
				ID:        "target2",
				Category:  "System",
				Label:     "Temp Files",
				Paths:     []string{"/test/temp"},
				Patterns:  []string{"*"},
				Risk:      RiskLow,
				TotalSize: 2048,
			},
		},
		TotalSize: 3072,
	}

	opts := CLIOptions{}

	// Should not panic
	RenderCLI(result, opts)
}

func TestPrintSelection(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", TotalSize: 100, Selected: true},
			{ID: "target2", TotalSize: 200, Selected: true},
			{ID: "target3", TotalSize: 300, Selected: false},
		},
		TotalSize: 600,
	}

	// printSelection writes to stdout, just verify it doesn't panic
	printSelection(result)
}

func TestPrintSelection_Empty(t *testing.T) {
	result := &ScanResult{
		Targets:   []*CleanTarget{},
		TotalSize: 0,
	}

	// Should not panic
	printSelection(result)
}

func TestSelectTargets(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", TotalSize: 100, Risk: RiskSafe, Selected: false},
			{ID: "target2", TotalSize: 200, Risk: RiskLow, Selected: false},
			{ID: "target3", TotalSize: 300, Risk: RiskModerate, Selected: false},
			{ID: "target4", TotalSize: 0, Risk: RiskSafe, Selected: false},
		},
	}

	selectTargets(result, true, false)

	if !result.Targets[0].Selected {
		t.Error("Expected target1 (Safe) to be selected")
	}
	if !result.Targets[1].Selected {
		t.Error("Expected target2 (Low) to be selected")
	}
	if result.Targets[2].Selected {
		t.Error("Expected target3 (Moderate) to NOT be selected")
	}
}

func TestSelectTargets_IncludeModerate(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", TotalSize: 100, Risk: RiskSafe, Selected: false},
			{ID: "target2", TotalSize: 200, Risk: RiskModerate, Selected: false},
		},
	}

	selectTargets(result, true, true)

	if !result.Targets[0].Selected {
		t.Error("Expected target1 to be selected")
	}
	if !result.Targets[1].Selected {
		t.Error("Expected target2 (Moderate) to be selected")
	}
}

func TestSelectTargets_Deselect(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", TotalSize: 100, Risk: RiskSafe, Selected: true},
			{ID: "target2", TotalSize: 200, Risk: RiskSafe, Selected: true},
		},
	}

	selectTargets(result, false, false)

	if result.Targets[0].Selected {
		t.Error("Expected target1 to be deselected")
	}
	if result.Targets[1].Selected {
		t.Error("Expected target2 to be deselected")
	}
}

func TestGetSelectedTargets(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", TotalSize: 100, Selected: true},
			{ID: "target2", TotalSize: 200, Selected: false},
			{ID: "target3", TotalSize: 300, Selected: true},
			{ID: "target4", TotalSize: 0, Selected: true},
		},
	}

	selected := getSelectedTargets(result)

	if len(selected) != 2 {
		t.Errorf("Expected 2 selected targets, got %d", len(selected))
	}
}

func TestGetSelectedTargets_None(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", TotalSize: 100, Selected: false},
			{ID: "target2", TotalSize: 200, Selected: false},
		},
	}

	selected := getSelectedTargets(result)

	if len(selected) != 0 {
		t.Errorf("Expected 0 selected targets, got %d", len(selected))
	}
}

func TestGetTotalSize(t *testing.T) {
	targets := []*CleanTarget{
		{TotalSize: 100},
		{TotalSize: 200},
		{TotalSize: 300},
	}

	total := getTotalSize(targets)
	if total != 600 {
		t.Errorf("Expected total 600, got %d", total)
	}
}

func TestGetTotalSize_Empty(t *testing.T) {
	targets := []*CleanTarget{}

	total := getTotalSize(targets)
	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{10240, "10.00 KB"},
		{1048576, "1.00 MB"},
		{1572864, "1.50 MB"},
		{104857600, "100.00 MB"},
		{1073741824, "1.00 GB"},
		{2147483648, "2.00 GB"},
		{1099511627776, "1.00 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatSize(%d) = %q, want %q", tt.size, result, tt.expected)
			}
		})
	}
}

func TestScanResult_Struct(t *testing.T) {
	result := &ScanResult{
		Targets: []*CleanTarget{
			{ID: "target1", TotalSize: 100},
		},
		TotalSize: 100,
	}

	if len(result.Targets) != 1 {
		t.Errorf("Expected 1 target, got %d", len(result.Targets))
	}
}

func TestScanResult_EmptyTargets(t *testing.T) {
	result := &ScanResult{
		Targets:   []*CleanTarget{},
		TotalSize: 0,
	}

	if len(result.Targets) != 0 {
		t.Errorf("Expected 0 targets, got %d", len(result.Targets))
	}
}
