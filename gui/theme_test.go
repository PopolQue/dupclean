package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestTheme_Color(t *testing.T) {
	th := NewDupCleanTheme()

	colors := []fyne.ThemeColorName{
		theme.ColorNamePrimary,
		theme.ColorNameBackground,
		theme.ColorNameMenuBackground,
		theme.ColorNameSelection,
		theme.ColorNameFocus,
		theme.ColorNameButton,
		theme.ColorNameInputBackground,
		theme.ColorNameScrollBar,
		theme.ColorNameSeparator,
		theme.ColorNameForeground, // Default case
	}

	for _, name := range colors {
		// Test Dark Variant
		dark := th.Color(name, theme.VariantDark)
		if dark == nil {
			t.Errorf("Dark color for %s should not be nil", name)
		}

		// Test Light Variant
		light := th.Color(name, theme.VariantLight)
		if light == nil {
			t.Errorf("Light color for %s should not be nil", name)
		}
	}
}

func TestTheme_IconAndFont(t *testing.T) {
	th := NewDupCleanTheme()
	if th.Icon(theme.IconNameHome) == nil {
		t.Error("Icon should not be nil")
	}
	if th.Font(fyne.TextStyle{}) == nil {
		t.Error("Font should not be nil")
	}
}

func TestTheme_Size(t *testing.T) {
	th := NewDupCleanTheme()

	tests := []struct {
		name fyne.ThemeSizeName
		want float32
	}{
		{theme.SizeNamePadding, 8},
		{theme.SizeNameInlineIcon, 20},
		{theme.SizeNameHeadingText, 24},
		{theme.SizeNameSubHeadingText, 18},
		{theme.SizeNameText, 14},
		{theme.SizeNameCaptionText, theme.DefaultTheme().Size(theme.SizeNameCaptionText)},
	}

	for _, tt := range tests {
		got := th.Size(tt.name)
		if got != tt.want {
			t.Errorf("Size(%s) = %f, want %f", tt.name, got, tt.want)
		}
	}
}
