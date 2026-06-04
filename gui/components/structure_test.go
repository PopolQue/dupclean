package components

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestSectionHeader(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	headerHoriz := SectionHeader("Title", "Subtitle", true)
	if headerHoriz == nil {
		t.Fatal("SectionHeader (horizontal) should not be nil")
	}
	// Verify horizontal structure contains accent and labels
	box := headerHoriz.(*fyne.Container)
	if len(box.Objects) != 2 {
		t.Errorf("Expected 2 objects, got %d", len(box.Objects))
	}

	headerVert := SectionHeader("Title", "Subtitle", false)
	if headerVert == nil {
		t.Fatal("SectionHeader (vertical) should not be nil")
	}
}

func TestActionFooter(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	footer := ActionFooter(widget.NewLabel("L"), widget.NewLabel("C"), widget.NewLabel("R"))
	if footer == nil {
		t.Fatal("ActionFooter should not be nil")
	}

	// Footer is HBox, should have 3 labels + 2 spacers = 5 objects
	box := footer.(*fyne.Container)
	if len(box.Objects) != 5 {
		t.Errorf("Expected 5 objects (3 items + 2 spacers), got %d", len(box.Objects))
	}
}

func TestButtons(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	pBtn := PrimaryButton("P", nil, nil)
	if pBtn == nil {
		t.Fatal("PrimaryButton should not be nil")
	}
	if pBtn.Importance != widget.HighImportance {
		t.Errorf("Expected HighImportance, got %v", pBtn.Importance)
	}

	sBtn := SecondaryButton("S", nil, nil)
	if sBtn == nil {
		t.Fatal("SecondaryButton should not be nil")
	}
	if sBtn.Importance != widget.LowImportance {
		t.Errorf("Expected LowImportance, got %v", sBtn.Importance)
	}
}
