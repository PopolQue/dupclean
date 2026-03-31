package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
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
			label := canvas.NewText("Item", theme.ForegroundColor())
			label.TextSize = 14
			label.TextStyle = fyne.TextStyle{Bold: true}

			return container.NewHBox(
				container.NewMax(icon),
				label,
			)
		},
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			if i < 0 || i >= len(data) {
				return
			}

			item := data[i]
			hbox := obj.(*fyne.Container)
			icon := hbox.Objects[0].(*fyne.Container).Objects[0].(*widget.Icon)
			label := hbox.Objects[1].(*canvas.Text)

			icon.SetResource(item.Icon)
			label.Text = item.Name

			// Highlight selected item
			if i == selected {
				label.Color = theme.Color(theme.ColorNamePrimary)
			} else {
				label.Color = theme.ForegroundColor()
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
	}

	list := Sidebar(items)

	// Wrap in a scroll container with fixed width
	scroll := container.NewScroll(list)
	scroll.SetMinSize(fyne.NewSize(200, 0))

	return scroll
}
