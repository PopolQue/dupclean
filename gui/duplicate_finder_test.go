package gui

import (
	"testing"
)

func TestDuplicateFinderWidgetPanic(t *testing.T) {
	state := &AppState{
		ProcessManager: &ProcessManager{},
	}
	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DuplicateFinderWidget panicked: %v", r)
		}
	}()
	DuplicateFinderWidget(state, "byte")
}
