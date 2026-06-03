package gui

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/PopolQue/dupclean/internal/version"
)

func TestShowChangelogIfNeeded(t *testing.T) {
	app := test.NewApp()
	w := test.NewWindow(nil)
	p := app.Preferences()

	// Initial run: version is empty, should trigger showChangelog (simulated)
	// We check if it sets the preference
	p.SetString(lastSeenVersionKey, "")

	ShowChangelogIfNeeded(w)

	if p.String(lastSeenVersionKey) != version.Version {
		t.Errorf("Expected lastSeenVersion to be updated to %s, got %s", version.Version, p.String(lastSeenVersionKey))
	}
}

func TestShowFullChangelog(t *testing.T) {
	w := test.NewWindow(nil)
	// Just check it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ShowFullChangelog panicked: %v", r)
		}
	}()
	ShowFullChangelog(w)
}
