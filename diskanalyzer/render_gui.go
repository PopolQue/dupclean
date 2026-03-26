package diskanalyzer

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// RenderGUI creates a table visualization of disk usage.
func RenderGUI(result *AnalysisResult) {
	fyneApp := app.New()
	w := fyneApp.NewWindow("DupClean - Disk Analyzer")

	// Get largest directories
	largestDirs := LargestDirs(result, 100)

	// Create table
	table := widget.NewTable(
		func() (int, int) {
			return len(largestDirs), 3 // rows, cols: Name, Size, Files
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			node := largestDirs[id.Row]
			switch id.Col {
			case 0:
				label.SetText(node.Name)
			case 1:
				label.SetText(formatSize(node.TotalSize))
			case 2:
				label.SetText(strconv.Itoa(len(node.Files)))
			}
		},
	)

	table.SetColumnWidth(0, 400)
	table.SetColumnWidth(1, 100)
	table.SetColumnWidth(2, 80)

	// Create header
	header := container.NewVBox(
		widget.NewRichTextFromMarkdown("**Disk Usage Analyzer**"),
		widget.NewLabel(fmt.Sprintf("Root: %s | Total: %s | Files: %d",
			result.Root.Path,
			formatSize(result.TotalSize),
			result.FileCount,
		)),
	)

	// Create instructions
	instructions := widget.NewLabel("Click on a row to see details")

	// Layout
	content := container.NewBorder(
		header,
		instructions,
		nil,
		nil,
		table,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
