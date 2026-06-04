package gui

import (
	"os"
	"testing"
)

// Main test function to ensure consistent mocking across all gui package tests.
func TestMain(m *testing.M) {
	// Default to a mock implementation that does nothing.
	// Individual tests can still override this if specific behavior is needed.
	oldTrash := moveToTrash
	moveToTrash = func(path string) error { return nil }

	code := m.Run()

	moveToTrash = oldTrash
	os.Exit(code)
}
