package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ResultCard creates a standardized card for result items (files, cache targets, etc.)
func ResultCard(title string, description fyne.CanvasObject, metadata string, selector *widget.Check, actions fyne.CanvasObject) *widget.Card {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	metaLabel := widget.NewLabel(metadata)
	metaLabel.TextStyle = fyne.TextStyle{Italic: true}
	metaLabel.Importance = widget.MediumImportance

	content := container.NewVBox(
		titleLabel,
	)
	if description != nil {
		content.Add(description)
	}
	content.Add(metaLabel)

	// Combine selector and actions on the right
	rightSide := container.NewHBox()
	if selector != nil {
		rightSide.Add(selector)
	}
	if actions != nil {
		rightSide.Add(actions)
	}

	cardContent := container.NewBorder(nil, nil, nil, rightSide, content)
	return widget.NewCard("", "", cardContent)
}

// StatusBadge creates a colored "pill" label for status indicators
func StatusBadge(text string, importance widget.Importance) fyne.CanvasObject {
	label := widget.NewLabelWithStyle(text, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	label.Importance = importance

	background := canvas.NewRectangle(theme.Color(theme.ColorNameSelection)) // Default pill-like background
	background.CornerRadius = 10

	// Adjust background color based on importance if needed, but Fyne's Importance on Label
	// already handles text color. For a true "pill" we might need a custom layout.
	// For now, let's keep it simple with a themed label.
	return container.NewMax(label)
}

// StatsRow creates a consistent way to display summary metrics
func StatsRow(items []string) fyne.CanvasObject {
	row := container.NewHBox()
	for i, item := range items {
		label := widget.NewLabel(item)
		label.TextStyle = fyne.TextStyle{Bold: true}
		row.Add(label)
		if i < len(items)-1 {
			row.Add(widget.NewLabel("|"))
		}
	}
	return row
}
