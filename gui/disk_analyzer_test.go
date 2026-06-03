package gui

import (
	"fmt"
	"testing"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
	"github.com/PopolQue/dupclean/diskanalyzer"
)

// ... (existing functions)

func TestHandleAnalysisResult(t *testing.T) {
	_ = test.NewApp()
	pm := NewProcessManager()
	comp := &diskAnalyzerComponents{
		analyzeBtn:    widget.NewButton("", nil),
		progressLabel: widget.NewLabel(""),
		progressBar:   widget.NewProgressBar(),
	}
	state := &DiskAnalyzerState{
		ProcessManager: pm,
		components:     comp,
		Window:         test.NewWindow(nil),
	}

	// Test error case
	err := fmt.Errorf("analysis failed")
	handleAnalysisResult(state, nil, err)

	if state.IsAnalyzing {
		t.Error("IsAnalyzing should be false")
	}
	if comp.progressLabel.Text != "Error: analysis failed" {
		t.Errorf("Progress label = %s, want 'Error: analysis failed'", comp.progressLabel.Text)
	}

	// Test success case
	result := &diskanalyzer.AnalysisResult{
		FileCount: 5,
		Root:      &diskanalyzer.DirNode{Name: "root", Path: "/"},
	}
	handleAnalysisResult(state, result, nil)

	if state.Result != result {
		t.Error("Result not set correctly")
	}
}

func TestPrepareDiskAnalysisUI(t *testing.T) {
	_ = test.NewApp()
	pm := NewProcessManager()
	comp := &diskAnalyzerComponents{
		analyzeBtn:    widget.NewButton("", nil),
		progressLabel: widget.NewLabel(""),
		progressBar:   widget.NewProgressBar(),
	}
	state := &DiskAnalyzerState{
		ProcessManager: pm,
		components:     comp,
	}

	prepareDiskAnalysisUI(state)

	if !state.IsAnalyzing {
		t.Error("IsAnalyzing should be true")
	}
	if !comp.analyzeBtn.Disabled() {
		t.Error("analyzeBtn should be disabled")
	}
	if comp.progressBar.Value != 0.5 {
		t.Errorf("ProgressBar value = %f, want 0.5", comp.progressBar.Value)
	}
}

func TestPrepareAnalysisViewData(t *testing.T) {
	result := &diskanalyzer.AnalysisResult{
		FileCount: 1,
		TotalSize: 100,
		Root:      &diskanalyzer.DirNode{Name: "root", Path: "/"},
	}

	data := prepareAnalysisViewData(result)

	// Since LargestDirs/TypeBreakdown are library functions, we test coordination
	if data.LargestDirs == nil {
		t.Error("LargestDirs should not be nil")
	}
	if data.TypeBreakdown == nil {
		t.Error("TypeBreakdown should not be nil")
	}
}

func TestNewDiskAnalyzerState(t *testing.T) {
	w := test.NewWindow(nil)
	pm := NewProcessManager()
	state := NewDiskAnalyzerState(w, pm)

	if state.Window != w {
		t.Error("Window not correctly set")
	}
	if state.ProcessManager != pm {
		t.Error("ProcessManager not correctly set")
	}
}

func TestUpdateContent(t *testing.T) {
	state := &DiskAnalyzerState{
		ContentContainer: container.NewStack(),
	}

	newContent := container.NewStack()
	state.updateContent(newContent)

	if len(state.ContentContainer.Objects) != 1 || state.ContentContainer.Objects[0] != newContent {
		t.Error("ContentContainer was not updated correctly")
	}
}

func TestValidateAnalysisPath(t *testing.T) {
	state := &DiskAnalyzerState{FolderPath: ""}
	if err := validateAnalysisPath(state); err == nil {
		t.Error("Expected error for empty folder path")
	}

	state.FolderPath = "/valid/path"
	if err := validateAnalysisPath(state); err != nil {
		t.Errorf("Expected no error for valid path, got %v", err)
	}
}

func TestDiskAnalyzerWidget_Init(t *testing.T) {
	_ = test.NewApp()
	state := &DiskAnalyzerState{
		ProcessManager: NewProcessManager(),
		Window:         test.NewWindow(nil),
	}
	// Verify it does not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DiskAnalyzerWidget panicked: %v", r)
		}
	}()
	DiskAnalyzerWidget(state)
}

func TestPerformDiskAnalysis(t *testing.T) {
	oldWalker := diskWalker
	defer func() { diskWalker = oldWalker }()

	// Mock the walker
	diskWalker = func(root string, opts diskanalyzer.WalkOptions) (*diskanalyzer.AnalysisResult, []error, error) {
		return &diskanalyzer.AnalysisResult{FileCount: 10}, nil, nil
	}

	state := &DiskAnalyzerState{FolderPath: "/test/path"}
	result, err := performDiskAnalysis(state)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result == nil || result.FileCount != 10 {
		t.Errorf("Unexpected result: %v", result)
	}
}
