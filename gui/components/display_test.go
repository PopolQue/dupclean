package components

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestResultCard(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	selector := widget.NewCheck("Check", nil)
	action := widget.NewButton("Action", nil)
	card := ResultCard("Title", widget.NewLabel("Desc"), "Metadata", selector, action)

	if card == nil {
		t.Fatal("ResultCard should not be nil")
	}

	// Verify content - cardContent is a Border layout in a Card
}

func TestStatusBadge(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	badge := StatusBadge("Safe", widget.SuccessImportance)

	if badge == nil {
		t.Fatal("StatusBadge should not be nil")
	}

	// Badge is a Stack container with a Label
	stack := badge.(*fyne.Container)
	if len(stack.Objects) != 1 {
		t.Errorf("Expected 1 object in badge, got %d", len(stack.Objects))
	}
	label := stack.Objects[0].(*widget.Label)
	if label.Text != "Safe" {
		t.Errorf("Expected text 'Safe', got %s", label.Text)
	}
}

func TestStatsRow(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	row := StatsRow([]string{"Item1", "Item2"})

	if row == nil {
		t.Fatal("StatsRow should not be nil")
	}

	// StatsRow is HBox with items and separator
	box := row.(*fyne.Container)
	if len(box.Objects) != 3 { // Item1, |, Item2
		t.Errorf("Expected 3 objects (2 items + 1 separator), got %d", len(box.Objects))
	}
}

func TestStatsRow_Empty(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	row := StatsRow([]string{})
	if row == nil {
		t.Fatal("StatsRow should not be nil")
	}
	// StatsRow for empty input should just be an HBox (possibly empty)
	box := row.(*fyne.Container)
	if len(box.Objects) != 0 {
		t.Errorf("Expected 0 objects, got %d", len(box.Objects))
	}
}
