package diskanalyzer

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/PopolQue/dupclean/internal/fsutil"
)

// CLIOptions configures the CLI renderer output.
type CLIOptions struct {
	TopN      int   // Show N largest files (0 = don't show)
	ByType    bool  // Show type breakdown instead of tree
	OlderThan int   // Show files older than N days (0 = don't show)
	MinSize   int64 // Minimum file size for old files query
	Depth     int   // Tree depth (default: 2)
}

// RenderCLI renders the AnalysisResult to the provided writer.
func RenderCLI(w io.Writer, result *AnalysisResult, opts CLIOptions) {
	if opts.Depth <= 0 {
		opts.Depth = 2
	}

	// Print header
	_, _ = fmt.Fprintf(w, "%s  —  %s\n", result.Root.Path, fsutil.FormatBytes(result.TotalSize))
	_, _ = fmt.Fprintln(w, strings.Repeat("━", 60))
	_, _ = fmt.Fprintln(w)

	if opts.ByType {
		renderByType(w, result)
		return
	}

	if opts.TopN > 0 {
		renderTopFiles(w, result, opts.TopN)
		_, _ = fmt.Fprintln(w)
	}

	if opts.OlderThan > 0 {
		renderOldFiles(w, result, opts.OlderThan, opts.MinSize)
		_, _ = fmt.Fprintln(w)
	}

	// Render tree
	renderTree(w, result.Root, 0, opts.Depth, result.TotalSize)
}

// renderTree prints a directory tree with bar charts.
func renderTree(w io.Writer, node *DirNode, depth, maxDepth int, totalSize int64) {
	if depth > maxDepth {
		return
	}

	// Simple tree indentation
	indent := strings.Repeat("  ", depth)

	// Collect children for sorting
	type item struct {
		name  string
		size  int64
		isDir bool
	}
	var items []item
	for _, f := range node.Files {
		items = append(items, item{name: f.Name, size: f.Size, isDir: false})
	}
	for _, d := range node.Children {
		items = append(items, item{name: d.Name + "/", size: d.TotalSize, isDir: true})
	}

	// Sort by size desc
	sort.Slice(items, func(i, j int) bool {
		return items[i].size > items[j].size
	})

	for _, item := range items {
		// Calculate percentage
		pct := float64(item.size) / float64(totalSize) * 100

		// Bar chart
		bar := makeBar(item.size, totalSize, 25)
		suffix := ""

		_, _ = fmt.Fprintf(w, "%s  %-20s  %s  %7s  %5.1f%%\n",
			indent,
			item.name+suffix,
			bar,
			fsutil.FormatBytes(item.size),
			pct)

		// Recurse into children
		if item.isDir {
			for _, d := range node.Children {
				if d.Name+"/" == item.name {
					renderTree(w, d, depth+1, maxDepth, totalSize)
					break
				}
			}
		}
	}
}

// renderTopFiles prints the N largest files.
func renderTopFiles(w io.Writer, result *AnalysisResult, n int) {
	files := TopFiles(result, n)

	_, _ = fmt.Fprintln(w, "Top largest files")
	_, _ = fmt.Fprintln(w, strings.Repeat("━", 60))

	for i, f := range files {
		_, _ = fmt.Fprintf(w, "   %2d   %8s   %s\n", i+1, fsutil.FormatBytes(f.Size), f.Path)
	}
}

// renderByType prints the type breakdown.
func renderByType(w io.Writer, result *AnalysisResult) {
	stats := TypeBreakdown(result)
	_, _ = fmt.Fprintln(w, "By file type")
	_, _ = fmt.Fprintln(w, strings.Repeat("━", 60))

	for _, s := range stats {
		bar := makeBar(s.TotalSize, result.TotalSize, 25)
		_, _ = fmt.Fprintf(w, "  %-8s  %s  %7s  %5.1f%%  (%d files)\n",
			s.Ext, bar, fsutil.FormatBytes(s.TotalSize), s.PctOfDisk, s.Count)
	}
}

// renderOldFiles prints files older than N days.
func renderOldFiles(w io.Writer, result *AnalysisResult, days int, minSize int64) {
	files := OldFiles(result, days, minSize)

	if len(files) == 0 {
		_, _ = fmt.Fprintf(w, "No files older than %d days (min size: %s)\n", days, fsutil.FormatBytes(minSize))
		return
	}

	_, _ = fmt.Fprintf(w, "Files older than %d days (min size: %s)\n", days, fsutil.FormatBytes(minSize))
	_, _ = fmt.Fprintln(w, strings.Repeat("━", 60))

	for i, f := range files {
		if i >= 20 {
			_, _ = fmt.Fprintf(w, "   ... and %d more\n", len(files)-i)
			break
		}
		_, _ = fmt.Fprintf(w, "   %2d   %8s   %s\n", i+1, fsutil.FormatBytes(f.Size), f.Path)
	}
}

// makeBar creates a text bar chart.
func makeBar(value, total int64, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}

	pct := float64(value) / float64(total)
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}
