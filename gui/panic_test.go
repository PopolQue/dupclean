package gui

import (
	"testing"
)

func TestCacheCleanerWidgetPanic(t *testing.T) {
	state := &CacheCleanerState{
		ProcessManager: &ProcessManager{},
	}
	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("CacheCleanerWidget panicked: %v", r)
		}
	}()
	CacheCleanerWidget(state)
}
