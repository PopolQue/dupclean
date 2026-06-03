package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ToolPage wraps a content in a standard tool layout with header and padding
func ToolPage(title, subtitle string, content fyne.CanvasObject) fyne.CanvasObject {
	return ToolPageWithFooter(title, subtitle, content, nil)
}

// ToolPageWithFooter wraps a content in a standard tool layout with a fixed header and footer
func ToolPageWithFooter(title, subtitle string, content fyne.CanvasObject, footer fyne.CanvasObject) fyne.CanvasObject {
	header := SectionHeader(title, subtitle, false)

	// Ensure content is scrolled if it's not already
	scrolledContent := content
	if _, ok := content.(*container.Scroll); !ok {
		// Only wrap if it's likely to need scrolling
		scrolledContent = container.NewVScroll(content)
	}

	main := container.NewBorder(header, footer, nil, nil, scrolledContent)
	return container.NewPadded(main)
}

// ToolHome creates a standard tool homepage layout with unified spacing and structure
func ToolHome(title, subtitle string, cards []fyne.CanvasObject, actionBtn *widget.Button, progress fyne.CanvasObject, extra fyne.CanvasObject) fyne.CanvasObject {
	body := container.NewVBox()

	// Add input cards
	for _, card := range cards {
		body.Add(card)
	}

	// Add main action button (centered)
	if actionBtn != nil {
		body.Add(container.NewHBox(layout.NewSpacer(), actionBtn, layout.NewSpacer()))
	}

	// Add progress indicator
	if progress != nil {
		body.Add(progress)
	}

	// Add extra elements (disclaimers, etc.)
	if extra != nil {
		body.Add(extra)
	}

	return ToolPage(title, subtitle, body)
}

// VerticalDivider creates a simple vertical line for separators
func VerticalDivider() fyne.CanvasObject {
	rect := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	rect.SetMinSize(fyne.NewSize(1, 24))
	return rect
}

// FixedTabLayout creates a rigid four-block layout: Header/Subheader, Options, Action Row, and Log/Menu
func FixedTabLayout(title, subheader string, options fyne.CanvasObject, actionBtn *widget.Button, progressBar *widget.ProgressBar, logMenu fyne.CanvasObject) fyne.CanvasObject {
	headerBlock := SectionHeader(title, subheader, true)

	// Block 2: OPTIONS (Scrollable)
	optionsBlock := container.NewVScroll(options)

	// Block 3: START BUTTON | PROGRESS BAR
	if progressBar != nil {
		progressBar.Min = 0
		progressBar.Max = 1
	}

	// Ensure button has some weight
	if actionBtn != nil {
		actionBtn.Importance = widget.HighImportance
	}

	actionRow := container.NewGridWithColumns(2)
	if actionBtn != nil {
		actionRow.Add(actionBtn)
	} else {
		actionRow.Add(layout.NewSpacer())
	}
	if progressBar != nil {
		actionRow.Add(progressBar)
	} else {
		actionRow.Add(layout.NewSpacer())
	}

	actionBlock := container.NewVBox(widget.NewSeparator(), actionRow, widget.NewSeparator())

	// Block 4: LOG/ACTION MENU
	logBlock := logMenu
	if logBlock == nil {
		logBlock = layout.NewSpacer()
	}

	// Compose the layout:
	// Header (Top)
	// Log/Menu (Bottom)
	// Middle (Action Row at bottom of center, Options in the rest)
	centerArea := container.NewBorder(nil, actionBlock, nil, nil, optionsBlock)
	main := container.NewBorder(headerBlock, logBlock, nil, nil, centerArea)

	return container.NewPadded(main)
}

// StatusPage creates a centered status page (success, error, empty state)
func StatusPage(title, subtitle string, icon fyne.Resource, message, subMessage string, actions fyne.CanvasObject) fyne.CanvasObject {
	iconImg := canvas.NewImageFromResource(icon)
	iconImg.FillMode = canvas.ImageFillContain
	iconImg.SetMinSize(fyne.NewSize(80, 80))

	messageLabel := widget.NewLabelWithStyle(message, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	messageLabel.Importance = widget.HighImportance
	messageLabel.SizeName = theme.SizeNameHeadingText

	subLabel := widget.NewLabel(subMessage)
	subLabel.TextStyle = fyne.TextStyle{Italic: true}
	subLabel.Alignment = fyne.TextAlignCenter

	content := container.NewCenter(container.NewVBox(
		layout.NewSpacer(),
		iconImg,
		messageLabel,
		subLabel,
		layout.NewSpacer(),
	))

	// For status pages, the actions are often centered but we'll use footer for consistency if provided
	var body fyne.CanvasObject = content
	if actions != nil {
		body = container.NewBorder(nil, container.NewCenter(actions), nil, nil, content)
	}

	return ToolPage(title, subtitle, body)
}

// ProgressPage creates a centered progress page for long-running operations
func ProgressPage(title, subtitle, cardTitle, cardSubtitle string, label *widget.Label, bar *widget.ProgressBar) fyne.CanvasObject {
	content := container.NewVBox(
		layout.NewSpacer(),
		widget.NewCard(cardTitle, cardSubtitle, container.NewVBox(
			label,
			bar,
		)),
		layout.NewSpacer(),
	)
	return ToolPage(title, subtitle, content)
}
