package components

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestSectionHeader(t *testing.T) {
	_ = test.NewApp()

	headerHoriz := SectionHeader("Title", "Subtitle", true)
	if headerHoriz == nil {
		t.Fatal("SectionHeader (horizontal) should not be nil")
	}

	headerVert := SectionHeader("Title", "Subtitle", false)
	if headerVert == nil {
		t.Fatal("SectionHeader (vertical) should not be nil")
	}
}

func TestActionFooter(t *testing.T) {
	_ = test.NewApp()

	footer := ActionFooter(widget.NewLabel("L"), widget.NewLabel("C"), widget.NewLabel("R"))
	if footer == nil {
		t.Fatal("ActionFooter should not be nil")
	}

	// Test nil components
	footerNil := ActionFooter(nil, nil, nil)
	if footerNil == nil {
		t.Fatal("ActionFooter (nil) should not be nil")
	}
}

func TestButtons(t *testing.T) {
	_ = test.NewApp()

	if PrimaryButton("P", nil, nil) == nil {
		t.Error("PrimaryButton should not be nil")
	}

	if SecondaryButton("S", nil, nil) == nil {
		t.Error("SecondaryButton should not be nil")
	}
}
