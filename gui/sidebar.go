package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SidebarItem represents a navigation item in the sidebar
type SidebarItem struct {
	Name    string
	Icon    fyne.Resource
	Tooltip string
	OnClick func()
}

// Sidebar creates a navigation sidebar with the given items
func Sidebar(items []SidebarItem) *widget.List {
	data := items
	selected := -1

	list := widget.NewList(
		func() int { return len(data) },
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.HomeIcon())
			label := widget.NewLabel("Item")
			label.TextStyle = fyne.TextStyle{Bold: true}

			return container.NewPadded(
				container.NewHBox(
					icon,
					label,
				),
			)
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			if i < 0 || i >= len(data) {
				return
			}

			item := data[i]
			padded := obj.(*fyne.Container)
			hbox := padded.Objects[0].(*fyne.Container)
			icon := hbox.Objects[0].(*widget.Icon)
			label := hbox.Objects[1].(*widget.Label)

			icon.SetResource(item.Icon)
			label.SetText(item.Name)

			// Highlight selected item
			if i == selected {
				label.Importance = widget.HighImportance
			} else {
				label.Importance = widget.MediumImportance
			}
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		selected = id
		list.Refresh()
	}

	return list
}

// CreateSidebar creates the main navigation sidebar
func CreateSidebar() fyne.CanvasObject {
	items := []SidebarItem{
		{
			Name: "Duplicate Finder",
			Icon: theme.SearchIcon(),
		},
		{
			Name: "Cache Cleaner",
			Icon: theme.DeleteIcon(),
		},
		{
			Name: "Disk Analyzer",
			Icon: theme.StorageIcon(),
		},
		{
			Name: "About",
			Icon: theme.InfoIcon(),
		},
	}

	list := Sidebar(items)

	// Wrap in a scroll container with fixed width
	scroll := container.NewScroll(list)
	scroll.SetMinSize(fyne.NewSize(200, 0))

	return scroll
}
