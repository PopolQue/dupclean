package gui

import (
	"testing"

	"fyne.io/fyne/v2/widget"
)

func TestProcessManager(t *testing.T) {
	pm := NewProcessManager()
	btn := widget.NewButton("Test", nil)

	// Test RegisterStartButton when not running
	pm.RegisterStartButton(btn)
	if btn.Disabled() {
		t.Error("Button should be enabled when not running")
	}

	// Test SetProcessRunning(true)
	pm.SetProcessRunning(true)
	if !btn.Disabled() {
		t.Error("Button should be disabled when running")
	}

	// Test RegisterStartButton when running
	btn2 := widget.NewButton("Test2", nil)
	pm.RegisterStartButton(btn2)
	if !btn2.Disabled() {
		t.Error("Button should be disabled when registered while running")
	}

	// Test SetProcessRunning(false)
	pm.SetProcessRunning(false)
	if btn.Disabled() {
		t.Error("Button should be enabled when not running")
	}
	if btn2.Disabled() {
		t.Error("Button2 should be enabled when not running")
	}
}
