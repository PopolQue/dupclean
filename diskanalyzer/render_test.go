package diskanalyzer

import (
	"testing"
)

// TestMakeBar_Full verifies makeBar fills correctly.
func TestMakeBar_Full(t *testing.T) {
	bar := makeBar(100, 100, 20)
	if len(bar) < 20 {
		t.Errorf("Expected full bar to be at least 20 chars, got %d", len(bar))
	}
}

// TestMakeBar_Empty verifies makeBar handles empty values.
func TestMakeBar_Empty(t *testing.T) {
	bar := makeBar(0, 100, 20)
	if len(bar) < 20 {
		t.Errorf("Expected empty bar to be at least 20 chars, got %d", len(bar))
	}
}
