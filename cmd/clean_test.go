package cmd

import (
	"testing"
)

func TestPrepareClean(t *testing.T) {
	opts, cliOpts, err := prepareClean("24h", 8, true, true, true)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if opts.Concurrency != 8 {
		t.Errorf("Expected concurrency 8, got %d", opts.Concurrency)
	}
	if !cliOpts.DryRun || !cliOpts.Permanent || !cliOpts.Yes {
		t.Error("CLI options not set correctly")
	}

	_, _, err = prepareClean("invalid", 0, false, false, false)
	if err == nil {
		t.Error("Expected error for invalid minAge")
	}
}

func TestRunClean(t *testing.T) {
	// Set test-friendly state
	cleanDryRun = true
	cleanYes = true
	cleanMinAge = ""
	cleanWorkers = 0
	cleanCategory = ""
	cleanTargetIDs = []string{}

	// Test: No targets found
	cleanCategory = "nonexistent"
	err := runClean()
	if err != nil {
		t.Errorf("runClean failed for empty targets: %v", err)
	}

	// Reset for next test
	cleanMinAge = ""

	// Test: successful scan (mocked via dry-run/yes)
	// This will call cleaner.Scan() if targets exist
	err = runClean()
	if err != nil {
		t.Errorf("runClean failed for scan: %v", err)
	}
}
