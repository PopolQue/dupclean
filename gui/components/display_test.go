package components

import (
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestResultCard(t *testing.T) {
	_ = test.NewApp()
	card := ResultCard("Title", widget.NewLabel("Desc"), "Metadata", widget.NewCheck("Check", nil), widget.NewButton("Action", nil))

	if card == nil {
		t.Fatal("ResultCard should not be nil")
	}
}

func TestStatusBadge(t *testing.T) {
	_ = test.NewApp()
	badge := StatusBadge("Safe", widget.SuccessImportance)

	if badge == nil {
		t.Fatal("StatusBadge should not be nil")
	}
}

func TestStatsRow(t *testing.T) {
	_ = test.NewApp()
	row := StatsRow([]string{"Item1", "Item2"})

	if row == nil {
		t.Fatal("StatsRow should not be nil")
	}
}
