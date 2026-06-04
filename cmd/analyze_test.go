package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecuteAnalyze(t *testing.T) {
	// Create a temp directory for analysis
	tmpDir := t.TempDir()

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)

	// Test invalid path
	err := executeAnalyze(cmd, "/nonexistent", buf)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}

	// Test valid path
	err = executeAnalyze(cmd, tmpDir, buf)
	if err != nil {
		t.Errorf("Expected no error for valid path, got %v", err)
	}
}

func TestRunAnalyze(t *testing.T) {
	tmpDir := t.TempDir()
	cmd := &cobra.Command{}

	err := runAnalyze(cmd, tmpDir)
	if err != nil {
		t.Errorf("runAnalyze failed: %v", err)
	}
}
