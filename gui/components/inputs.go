package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// FolderPicker creates a combined entry and browse button component
func FolderPicker(placeholder string, initialPath string, showEntry bool, window fyne.Window, onSelected func(string)) fyne.CanvasObject {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(placeholder)
	entry.SetText(initialPath)

	browseBtn := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(dir fyne.ListableURI, err error) {
			if err != nil || dir == nil {
				return
			}
			path := dir.Path()
			entry.SetText(path)
			if onSelected != nil {
				onSelected(path)
			}
		}, window)
	})

	// Also update on manual entry change
	entry.OnChanged = func(s string) {
		if onSelected != nil {
			onSelected(s)
		}
	}

	if showEntry {
		return container.NewBorder(nil, nil, nil, browseBtn, entry)
	}
	return browseBtn
}
