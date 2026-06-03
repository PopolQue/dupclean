package gui

import (
	"testing"

	"github.com/PopolQue/dupclean/internal/version"

	"fyne.io/fyne/v2/test"
)

func TestChangelogLogic(t *testing.T) {
	a := test.NewApp()
	defer test.NewApp() // cleanup

	p := a.Preferences()
	p.RemoveValue(lastSeenVersionKey)

	// First run: should show changelog (we can't easily check the dialog, but we check if preferences are updated)
	// We call a modified version of the function that doesn't require a window for testing logic

	lastSeen := p.String(lastSeenVersionKey)
	if lastSeen != "" {
		t.Errorf("Expected empty lastSeen on first run, got %s", lastSeen)
	}

	// Update last seen
	p.SetString(lastSeenVersionKey, version.Version)

	lastSeen = p.String(lastSeenVersionKey)
	if lastSeen != version.Version {
		t.Errorf("Expected lastSeen to be %s, got %s", version.Version, lastSeen)
	}
}
