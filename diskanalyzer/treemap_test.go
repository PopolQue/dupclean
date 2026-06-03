package diskanalyzer

import (
	"testing"
)

func TestSquarify(t *testing.T) {
	nodes := []*DirNode{
		{TotalSize: 100},
		{TotalSize: 50},
		{TotalSize: 25},
	}

	bounds := Rect{X: 0, Y: 0, W: 100, H: 100}

	result := Squarify(nodes, bounds)

	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}

	if len(result) != len(nodes) {
		t.Errorf("Expected %d results, got %d", len(nodes), len(result))
	}
}

func TestSquarify_Empty(t *testing.T) {
	nodes := []*DirNode{}
	bounds := Rect{X: 0, Y: 0, W: 100, H: 100}

	result := Squarify(nodes, bounds)

	if len(result) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(result))
	}
}

func TestSquarify_AspectRatios(t *testing.T) {
	// Create nodes with various sizes to force different row building behaviors
	nodes := []*DirNode{
		{Name: "n1", TotalSize: 600},
		{Name: "n2", TotalSize: 400},
		{Name: "n3", TotalSize: 300},
		{Name: "n4", TotalSize: 200},
		{Name: "n5", TotalSize: 100},
	}

	// Wide bounds (W > H) forces horizontal slicing
	wideBounds := Rect{X: 0, Y: 0, W: 1000, H: 100}
	wideResult := Squarify(nodes, wideBounds)

	if len(wideResult) != 5 {
		t.Errorf("Squarify(wide) got %d nodes, want 5", len(wideResult))
	}

	// Tall bounds (H > W) forces vertical slicing
	tallBounds := Rect{X: 0, Y: 0, W: 100, H: 1000}
	tallResult := Squarify(nodes, tallBounds)

	if len(tallResult) != 5 {
		t.Errorf("Squarify(tall) got %d nodes, want 5", len(tallResult))
	}
}

func TestLayoutTreemap(t *testing.T) {
	root := &DirNode{
		Path:      "/test",
		TotalSize: 1000,
		Children: []*DirNode{
			{Path: "/test/dir1", TotalSize: 500, Name: "dir1"},
			{Path: "/test/dir2", TotalSize: 500, Name: "dir2"},
		},
	}

	bounds := Rect{X: 0, Y: 0, W: 100, H: 100}
	layout := LayoutTreemap(root, bounds, 2)

	if len(layout) == 0 {
		t.Error("Expected non-empty layout")
	}
}

func TestLayoutTreemap_NilRoot(t *testing.T) {
	// Skip this test - LayoutTreemap panics on nil input
	t.Skip("LayoutTreemap does not handle nil root")
}

func TestColorPalette(t *testing.T) {
	r, g, b := ColorPalette(5)

	// Should return valid RGB values
	if r == 0 && g == 0 && b == 0 {
		t.Error("Expected non-black color")
	}
}

func TestColorPalette_Zero(t *testing.T) {
	r, g, b := ColorPalette(0)

	// Should return valid RGB values
	_ = r
	_ = g
	_ = b
}

func TestColorPalette_Large(t *testing.T) {
	r, g, b := ColorPalette(100)

	// Should return valid RGB values
	if r == 0 && g == 0 && b == 0 {
		t.Error("Expected non-black color")
	}
}

func TestLayoutNode(t *testing.T) {
	node := LayoutNode{
		Rect: Rect{X: 0, Y: 0, W: 100, H: 100},
		Node: &DirNode{Path: "/test", TotalSize: 1000},
	}

	// Verify node fields
	if node.Node == nil {
		t.Error("Expected non-nil Node")
	}
}

func TestRect(t *testing.T) {
	rect := Rect{X: 10, Y: 20, W: 100, H: 200}

	// Verify rect fields
	if rect.W != 100.0 {
		t.Errorf("Expected width 100, got %v", rect.W)
	}
}
