package components

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// SectionHeader creates a consistent header for each tool section
func SectionHeader(title, subtitle string, horizontal bool) fyne.CanvasObject {
	titleLabel := widget.NewLabel(title)
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Importance = widget.HighImportance
	titleLabel.SizeName = theme.SizeNameHeadingText

	subtitleLabel := widget.NewLabel(subtitle)
	subtitleLabel.TextStyle = fyne.TextStyle{Italic: true}

	accent := canvas.NewRectangle(theme.Color(theme.ColorNamePrimary))
	accent.SetMinSize(fyne.NewSize(4, 32))
	accent.CornerRadius = 2

	if horizontal {
		return container.NewVBox(
			container.NewHBox(
				accent,
				titleLabel,
				layout.NewSpacer(),
				subtitleLabel,
			),
			widget.NewSeparator(),
		)
	}

	return container.NewVBox(
		container.NewHBox(
			accent,
			titleLabel,
			layout.NewSpacer(),
		),
		subtitleLabel,
		widget.NewSeparator(),
	)
}

// ActionFooter creates a standardized footer for results pages
func ActionFooter(left, center, right fyne.CanvasObject) fyne.CanvasObject {
	footer := container.NewHBox()
	if left != nil {
		footer.Add(left)
	}
	footer.Add(layout.NewSpacer())
	if center != nil {
		footer.Add(center)
		footer.Add(layout.NewSpacer())
	}
	if right != nil {
		footer.Add(right)
	}
	return footer
}

// PrimaryButton creates a button with high importance
func PrimaryButton(label string, icon fyne.Resource, tapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, tapped)
	btn.Importance = widget.HighImportance
	return btn
}

// SecondaryButton creates a button with low importance
func SecondaryButton(label string, icon fyne.Resource, tapped func()) *widget.Button {
	btn := widget.NewButtonWithIcon(label, icon, tapped)
	btn.Importance = widget.LowImportance
	return btn
}
