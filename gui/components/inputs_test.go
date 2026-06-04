package components

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestFolderPicker(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := test.NewWindow(nil)
	var selectedPath string
	onSelected := func(path string) {
		selectedPath = path
	}

	picker := FolderPicker("Pick a folder", "/tmp", true, w, onSelected)

	if picker == nil {
		t.Fatal("FolderPicker should not be nil")
	}

	// FolderPicker returns a Border container when showEntry is true
	border := picker.(*fyne.Container)

	// In NewBorder(nil, nil, nil, browseBtn, entry),
	// entry is the center (0), browseBtn is right (1)
	entry := border.Objects[0].(*widget.Entry)
	if entry.Text != "/tmp" {
		t.Errorf("Expected entry text '/tmp', got %s", entry.Text)
	}

	// Verify OnChanged triggers callback
	entry.SetText("/new/path")
	if selectedPath != "/new/path" {
		t.Errorf("Expected onSelected to be called with '/new/path', got %s", selectedPath)
	}
}

func TestFolderPicker_NoEntry(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	w := test.NewWindow(nil)
	picker := FolderPicker("Pick a folder", "/tmp", false, w, nil)

	if picker == nil {
		t.Fatal("FolderPicker should not be nil")
	}

	// Should return just a button
	if _, ok := picker.(*widget.Button); !ok {
		t.Errorf("Expected *widget.Button, got %T", picker)
	}
}
