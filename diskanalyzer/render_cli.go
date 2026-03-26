package diskanalyzer

import (
	"fmt"
	"strings"
)

// CLIOptions configures the CLI renderer output.
type CLIOptions struct {
	TopN      int  // Show N largest files (0 = don't show)
	ByType    bool // Show type breakdown instead of tree
	OlderThan int  // Show files older than N days (0 = don't show)
	MinSize   int64 // Minimum file size for old files query
	Depth     int  // Tree depth (default: 2)
}

// RenderCLI renders the AnalysisResult to stdout.
func RenderCLI(result *AnalysisResult, opts CLIOptions) {
	if opts.Depth <= 0 {
		opts.Depth = 2
	}

	// Print header
	fmt.Printf("%s  —  %s\n", result.Root.Path, formatSize(result.TotalSize))
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()

	if opts.ByType {
		renderByType(result)
		return
	}

	if opts.TopN > 0 {
		renderTopFiles(result, opts.TopN)
		fmt.Println()
	}

	if opts.OlderThan > 0 {
		renderOldFiles(result, opts.OlderThan, opts.MinSize)
		fmt.Println()
	}

	// Render tree
	renderTree(result.Root, 0, opts.Depth, result.TotalSize)
}

// renderTree prints a directory tree with bar charts.
func renderTree(node *DirNode, depth int, maxDepth int, totalSize int64) {
	if depth > maxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)

	// Combine files and children for sorting
	allItems := make([]struct {
		name string
		size int64
		isDir bool
	}, 0, len(node.Files)+len(node.Children))

	for _, f := range node.Files {
		allItems = append(allItems, struct {
			name  string
			size  int64
			isDir bool
		}{name: f.Name, size: f.Size, isDir: false})
	}

	for _, c := range node.Children {
		allItems = append(allItems, struct {
			name  string
			size  int64
			isDir bool
		}{name: c.Name, size: c.TotalSize, isDir: true})
	}

	// Sort by size descending
	for i := 0; i < len(allItems)-1; i++ {
		for j := i + 1; j < len(allItems); j++ {
			if allItems[j].size > allItems[i].size {
				allItems[i], allItems[j] = allItems[j], allItems[i]
			}
		}
	}

	// Print items
	for i, item := range allItems {
		if depth >= maxDepth && item.isDir && i >= 5 {
			remaining := len(allItems) - i
			fmt.Printf("%s  ... %d more folders\n", indent, remaining)
			break
		}

		bar := makeBar(item.size, totalSize, 25)
		pct := 0.0
		if totalSize > 0 {
			pct = float64(item.size) / float64(totalSize) * 100
		}

		suffix := ""
		if item.isDir {
			suffix = "/"
		}

		fmt.Printf("%s  %-20s  %s  %7s  %5.1f%%\n",
			indent,
			item.name+suffix,
			bar,
			formatSize(item.size),
			pct)

		// Recurse into children
		if item.isDir {
			for _, c := range node.Children {
				if c.Name == item.name {
					renderTree(c, depth+1, maxDepth, totalSize)
					break
				}
			}
		}
	}
}

// renderTopFiles prints the N largest files.
func renderTopFiles(result *AnalysisResult, n int) {
	files := TopFiles(result, n)

	fmt.Println("Top largest files")
	fmt.Println(strings.Repeat("━", 60))

	for i, f := range files {
		fmt.Printf("   %2d   %8s   %s\n", i+1, formatSize(f.Size), f.Path)
	}
}

// renderByType prints the type breakdown.
func renderByType(result *AnalysisResult) {
	stats := TypeBreakdown(result)

	fmt.Println("By file type")
	fmt.Println(strings.Repeat("━", 60))

	for _, s := range stats {
		bar := makeBar(s.TotalSize, result.TotalSize, 25)
		fmt.Printf("  %-8s  %s  %7s  %5.1f%%  (%d files)\n",
			s.Ext, bar, formatSize(s.TotalSize), s.PctOfDisk, s.Count)
	}
}

// renderOldFiles prints files older than N days.
func renderOldFiles(result *AnalysisResult, days int, minSize int64) {
	files := OldFiles(result, days, minSize)

	if len(files) == 0 {
		fmt.Printf("No files older than %d days (min size: %s)\n", days, formatSize(minSize))
		return
	}

	fmt.Printf("Files older than %d days (min size: %s)\n", days, formatSize(minSize))
	fmt.Println(strings.Repeat("━", 60))

	for i, f := range files {
		if i >= 20 {
			fmt.Printf("   ... and %d more\n", len(files)-i)
			break
		}
		fmt.Printf("   %2d   %8s   %s\n", i+1, formatSize(f.Size), f.Path)
	}
}

// makeBar creates a text bar chart.
func makeBar(value, total int64, width int) string {
	filled := 0
	if total > 0 {
		filled = int(float64(value) / float64(total) * float64(width))
	}
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

// formatSize returns a human-readable size string.
func formatSize(n int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case n >= TB:
		return fmt.Sprintf("%7.2f TB", float64(n)/float64(TB))
	case n >= GB:
		return fmt.Sprintf("%7.2f GB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%7.2f MB", float64(n)/float64(MB))
	case n >= KB:
		return fmt.Sprintf("%7.2f KB", float64(n)/float64(KB))
	default:
		return fmt.Sprintf("%7d B", n)
	}
}

// GetTerminalWidth returns the terminal width or a default.
func GetTerminalWidth(defaultWidth int) int {
	// TODO: Implement using golang.org/x/term
	return defaultWidth
}
