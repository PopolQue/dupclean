package components

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestFolderPicker(t *testing.T) {
	_ = test.NewApp()

	// Create a dummy window
	w := test.NewWindow(nil)

	picker := FolderPicker("Pick a folder", "/tmp", true, w, nil)

	if picker == nil {
		t.Fatal("FolderPicker should not be nil")
	}
}
