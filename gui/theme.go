package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type dupCleanTheme struct{}

func (m *dupCleanTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// For this rebrand, we'll focus on a modern dark aesthetic based on the logo
	if variant == theme.VariantDark {
		switch name {
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0xa8, G: 0x55, B: 0xf7, A: 0xff} // #A855F7 - Purple
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0x0d, G: 0x0d, B: 0x12, A: 0xff} // #0D0D12 - Deep Dark
		case theme.ColorNameMenuBackground:
			return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x24, A: 0xff} // Slightly lighter dark for menus
		case theme.ColorNameSelection:
			return color.NRGBA{R: 0xec, G: 0x48, B: 0x99, A: 0x50} // #EC4899 - Pink (with alpha)
		case theme.ColorNameFocus:
			return color.NRGBA{R: 0xa8, G: 0x55, B: 0xf7, A: 0x7f} // Purple focus
		case theme.ColorNameButton:
			return color.NRGBA{R: 0x2a, G: 0x2b, B: 0x60, A: 0xff} // #2A2B60 - Indigo for buttons
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 0x1a, G: 0x1b, B: 0x26, A: 0xff}
		case theme.ColorNameScrollBar:
			return color.NRGBA{R: 0xa8, G: 0x55, B: 0xf7, A: 0x99}
		case theme.ColorNameSeparator:
			return color.NRGBA{R: 0x2a, G: 0x2b, B: 0x60, A: 0xff}
		}
	} else {
		// Light theme fallback or adaptation
		switch name {
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0xa8, G: 0x55, B: 0xf7, A: 0xff}
		}
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m *dupCleanTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *dupCleanTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *dupCleanTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNamePadding {
		return 6
	}
	if name == theme.SizeNameInlineIcon {
		return 20
	}
	return theme.DefaultTheme().Size(name)
}

func NewDupCleanTheme() fyne.Theme {
	return &dupCleanTheme{}
}
