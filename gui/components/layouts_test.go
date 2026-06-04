package components

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestToolPage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	content := widget.NewLabel("Content")
	page := ToolPage("Title", "Sub", content)

	if page == nil {
		t.Fatal("ToolPage returned nil")
	}

	// ToolPage returns a Padded container
	padded := page.(*fyne.Container)
	if padded == nil {
		t.Fatal("Expected Padded container")
	}
}

func TestToolPageWithFooter(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	content := widget.NewLabel("Content")
	footer := widget.NewLabel("Footer")
	if ToolPageWithFooter("Title", "Sub", content, footer) == nil {
		t.Error("ToolPageWithFooter returned nil")
	}
}

func TestToolHome(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	if ToolHome("Title", "Sub", []fyne.CanvasObject{widget.NewLabel("Card")}, nil, nil, nil) == nil {
		t.Error("ToolHome returned nil")
	}
}

func TestFixedTabLayout(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	content := widget.NewLabel("Content")
	if FixedTabLayout("Title", "Sub", content, nil, nil, nil) == nil {
		t.Error("FixedTabLayout returned nil")
	}
}

func TestStatusPage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	if StatusPage("Title", "Sub", nil, "Msg", "SubMsg", nil) == nil {
		t.Error("StatusPage returned nil")
	}
}

func TestProgressPage(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	if ProgressPage("Title", "Sub", "CardTitle", "CardSub", widget.NewLabel("Label"), widget.NewProgressBar()) == nil {
		t.Error("ProgressPage returned nil")
	}
}

func TestVerticalDivider(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	divider := VerticalDivider()
	if divider == nil {
		t.Fatal("VerticalDivider returned nil")
	}

	rect := divider.(*canvas.Rectangle)
	if rect.MinSize().Width != 1 || rect.MinSize().Height != 24 {
		t.Errorf("Unexpected size: %v", rect.MinSize())
	}
}
