package diskanalyzer

import (
	"math"
	"sort"
)

// Rect represents a rectangle in the treemap.
type Rect struct {
	X, Y, W, H float64
}

// LayoutNode represents a DirNode with its assigned rectangle.
type LayoutNode struct {
	Node *DirNode
	Rect Rect
}

// Squarify performs the squarified treemap layout algorithm.
// Returns a list of LayoutNodes with computed rectangles.
func Squarify(nodes []*DirNode, bounds Rect) []LayoutNode {
	if len(nodes) == 0 {
		return []LayoutNode{}
	}

	// Sort nodes by size descending
	sorted := make([]*DirNode, len(nodes))
	copy(sorted, nodes)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalSize > sorted[j].TotalSize
	})

	// Calculate total size for normalization
	var totalSize int64
	for _, n := range sorted {
		totalSize += n.TotalSize
	}

	if totalSize == 0 {
		return []LayoutNode{}
	}

	// Run squarify algorithm
	rows := squarify(sorted, bounds, totalSize)

	// Convert to LayoutNodes
	result := make([]LayoutNode, 0)
	for _, row := range rows {
		for _, node := range row {
			result = append(result, node)
		}
	}

	return result
}

// squarify implements the core algorithm.
func squarify(nodes []*DirNode, bounds Rect, totalSize int64) [][]LayoutNode {
	if len(nodes) == 0 {
		return [][]LayoutNode{}
	}

	var rows [][]LayoutNode
	currentRow := make([]LayoutNode, 0)
	currentRowSize := int64(0)

	// Calculate aspect ratio of current row
	worstAspectRatio := func(row []LayoutNode, total int64) float64 {
		if len(row) == 0 || total == 0 {
			return math.MaxFloat64
		}

		// Calculate dimensions based on whether we're filling horizontally or vertically
		var side1, side2 float64
		if bounds.W >= bounds.H {
			side1 = bounds.H
			side2 = float64(total) / float64(totalSize) * bounds.W
		} else {
			side1 = bounds.W
			side2 = float64(total) / float64(totalSize) * bounds.H
		}

		if side1 == 0 || side2 == 0 {
			return math.MaxFloat64
		}

		// Find worst aspect ratio in the row
		worst := math.MaxFloat64
		used := 0.0
		for _, node := range row {
			used += float64(node.Node.TotalSize) / float64(totalSize)
			var nodeW, nodeH float64
			if bounds.W >= bounds.H {
				nodeW = side2 * (float64(node.Node.TotalSize) / float64(totalSize) / used)
				nodeH = side1
			} else {
				nodeW = side1
				nodeH = side2 * (float64(node.Node.TotalSize) / float64(totalSize) / used)
			}

			ratio := math.Max(nodeW/nodeH, nodeH/nodeW)
			if ratio < worst {
				worst = ratio
			}
		}

		return worst
	}

	// Position calculator
	positionRow := func(row []LayoutNode, bounds Rect, total int64) []LayoutNode {
		if len(row) == 0 {
			return row
		}

		var positioned []LayoutNode
		var offset float64

		if bounds.W >= bounds.H {
			// Horizontal layout
			width := bounds.H
			_ = width // unused for now

			for _, node := range row {
				height := float64(node.Node.TotalSize) / float64(total) * width
				positioned = append(positioned, LayoutNode{
					Node: node.Node,
					Rect: Rect{
						X: bounds.X + offset,
						Y: bounds.Y,
						W: height,
						H: width,
					},
				})
				offset += height
			}
		} else {
			// Vertical layout
			height := bounds.W
			_ = height // unused for now

			for _, node := range row {
				width := float64(node.Node.TotalSize) / float64(total) * height
				positioned = append(positioned, LayoutNode{
					Node: node.Node,
					Rect: Rect{
						X: bounds.X,
						Y: bounds.Y + offset,
						W: width,
						H: height,
					},
				})
				offset += width
			}
		}

		return positioned
	}

	// Main loop
	remainingBounds := bounds
	remainingNodes := nodes

	for len(remainingNodes) > 0 {
		node := remainingNodes[0]
		newRow := append(currentRow, LayoutNode{Node: node})
		newSize := currentRowSize + node.TotalSize

		// Check if adding this node worsens the aspect ratio
		currentWorst := worstAspectRatio(currentRow, currentRowSize)
		newWorst := worstAspectRatio(newRow, newSize)

		if len(currentRow) > 0 && newWorst >= currentWorst {
			// Current row is better, finalize it
			positioned := positionRow(currentRow, remainingBounds, currentRowSize)
			rows = append(rows, positioned)

			// Update bounds for next row
			if remainingBounds.W >= remainingBounds.H {
				usedWidth := float64(currentRowSize) / float64(totalSize) * remainingBounds.W
				remainingBounds.X += usedWidth
				remainingBounds.W -= usedWidth
			} else {
				usedHeight := float64(currentRowSize) / float64(totalSize) * remainingBounds.H
				remainingBounds.Y += usedHeight
				remainingBounds.H -= usedHeight
			}

			// Reset row
			currentRow = make([]LayoutNode, 0)
			currentRowSize = 0
		} else {
			// Add node to current row
			currentRow = newRow
			currentRowSize = newSize
			remainingNodes = remainingNodes[1:]
		}
	}

	// Finalize last row
	if len(currentRow) > 0 {
		positioned := positionRow(currentRow, remainingBounds, currentRowSize)
		rows = append(rows, positioned)
	}

	return rows
}

// LayoutTreemap recursively lays out a directory tree as a treemap.
// Returns all layout nodes including children.
func LayoutTreemap(root *DirNode, bounds Rect, maxDepth int) []LayoutNode {
	if maxDepth < 0 {
		return []LayoutNode{}
	}

	// Layout children at this level
	children := make([]*DirNode, 0, len(root.Children))
	for _, child := range root.Children {
		if child.TotalSize > 0 {
			children = append(children, child)
		}
	}

	layouts := Squarify(children, bounds)

	// Recurse into children if depth allows
	if maxDepth > 0 {
		for i, layout := range layouts {
			childBounds := layout.Rect
			// Add padding
			padding := 2.0
			childBounds.X += padding
			childBounds.Y += padding
			childBounds.W -= padding * 2
			childBounds.H -= padding * 2

			if childBounds.W > 10 && childBounds.H > 10 {
				childLayouts := LayoutTreemap(layout.Node, childBounds, maxDepth-1)
				layouts = append(layouts[:i+1], append(childLayouts, layouts[i+1:]...)...)
			}
		}
	}

	return layouts
}

// ColorPalette returns a color for a node based on depth.
func ColorPalette(depth int) (r, g, b uint8) {
	colors := []struct{ r, g, b uint8 }{
		{66, 133, 244},  // Blue
		{234, 67, 53},   // Red
		{251, 188, 5},   // Yellow
		{52, 168, 83},   // Green
		{155, 89, 182},  // Purple
		{26, 188, 156},  // Teal
		{230, 126, 34},  // Orange
	}

	c := colors[depth%len(colors)]
	return c.r, c.g, c.b
}
