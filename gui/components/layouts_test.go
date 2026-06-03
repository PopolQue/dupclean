package components

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestLayouts(t *testing.T) {
	_ = test.NewApp()
	content := widget.NewLabel("Content")

	if ToolPage("Title", "Sub", content) == nil {
		t.Error("ToolPage returned nil")
	}

	if ToolPageWithFooter("Title", "Sub", content, widget.NewLabel("Footer")) == nil {
		t.Error("ToolPageWithFooter returned nil")
	}

	if ToolHome("Title", "Sub", []fyne.CanvasObject{widget.NewLabel("Card")}, nil, nil, nil) == nil {
		t.Error("ToolHome returned nil")
	}

	if VerticalDivider() == nil {
		t.Error("VerticalDivider returned nil")
	}

	if FixedTabLayout("Title", "Sub", content, nil, nil, nil) == nil {
		t.Error("FixedTabLayout returned nil")
	}

	if StatusPage("Title", "Sub", nil, "Msg", "SubMsg", nil) == nil {
		t.Error("StatusPage returned nil")
	}

	if ProgressPage("Title", "Sub", "CardTitle", "CardSub", widget.NewLabel("Label"), widget.NewProgressBar()) == nil {
		t.Error("ProgressPage returned nil")
	}
}
