package cmd

import (
	"testing"

	"github.com/PopolQue/dupclean/scanner"
	"github.com/spf13/cobra"
)

func TestExecute(t *testing.T) {
	rootCmd.SetArgs([]string{"--help"})
	Execute()
}

func TestRunDuplicateFinder(t *testing.T) {
	oldRun := interactiveRun
	interactiveRun = func(g []scanner.DuplicateGroup, s scanner.ScanStats) {}
	defer func() { interactiveRun = oldRun }()

	cmd := &cobra.Command{}
	// Create a tmp dir for valid folder
	tmpDir := t.TempDir()

	// This will not run interactive.Run because we mocked it
	runDuplicateFinder(cmd, tmpDir)
}

func TestExecuteDuplicateFinder_InvalidPath(t *testing.T) {
	//...

	cmd := &cobra.Command{}
	_, _, err := executeDuplicateFinder(cmd, "/nonexistent", "byte", 90)
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}
