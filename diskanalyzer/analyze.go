package diskanalyzer

import (
	"slices"
	"time"
)

// TypeStat represents aggregated statistics for a file extension.
type TypeStat struct {
	Ext       string  `json:"ext"`
	TotalSize int64   `json:"total_size"`
	Count     int     `json:"count"`
	PctOfDisk float64 `json:"pct_of_disk"` // relative to total size
}

// TopFiles returns the N largest files across the entire tree.
func TopFiles(result *AnalysisResult, n int) []FileEntry {
	if n <= 0 || n > len(result.AllFiles) {
		n = len(result.AllFiles)
	}
	return result.AllFiles[:n]
}

// TypeBreakdown groups total bytes by file extension, sorted by size descending.
func TypeBreakdown(result *AnalysisResult) []TypeStat {
	// Aggregate by extension
	extStats := make(map[string]struct {
		size  int64
		count int
	})

	for _, f := range result.AllFiles {
		ext := f.Ext
		if ext == "" {
			ext = "(none)"
		}
		stats := extStats[ext]
		stats.size += f.Size
		stats.count++
		extStats[ext] = stats
	}

	// Convert to slice
	stats := make([]TypeStat, 0, len(extStats))
	for ext, s := range extStats {
		pct := 0.0
		if result.TotalSize > 0 {
			pct = float64(s.size) / float64(result.TotalSize) * 100
		}
		stats = append(stats, TypeStat{
			Ext:       ext,
			TotalSize: s.size,
			Count:     s.count,
			PctOfDisk: pct,
		})
	}

	// Sort by size descending
	slices.SortFunc(stats, func(a, b TypeStat) int {
		if b.TotalSize != a.TotalSize {
			return int(b.TotalSize - a.TotalSize)
		}
		return 0
	})

	return stats
}

// OldFiles returns files not modified in more than `days` days,
// filtered by minimum size, sorted by size descending.
func OldFiles(result *AnalysisResult, days int, minSize int64) []FileEntry {
	cutoff := time.Now().AddDate(0, 0, -days)
	var out []FileEntry

	for _, f := range result.AllFiles {
		if f.ModTime.Before(cutoff) && f.Size >= minSize {
			out = append(out, f)
		}
	}

	return out
}

// LargestDirs returns the N directories with the highest TotalSize.
func LargestDirs(result *AnalysisResult, n int) []*DirNode {
	var all []*DirNode

	// BFS to collect all nodes
	var bfs func(*DirNode)
	bfs = func(node *DirNode) {
		all = append(all, node)
		for _, child := range node.Children {
			bfs(child)
		}
	}
	bfs(result.Root)

	// Sort by TotalSize descending
	slices.SortFunc(all, func(a, b *DirNode) int {
		if b.TotalSize != a.TotalSize {
			return int(b.TotalSize - a.TotalSize)
		}
		return 0
	})

	if n > len(all) {
		n = len(all)
	}
	if n <= 0 {
		return all
	}
	return all[:n]
}

// FindPathToRoot returns the path from a node up to the root.
// Useful for showing "how did we get here?"
func FindPathToRoot(node *DirNode) []*DirNode {
	var path []*DirNode
	for node != nil {
		path = append(path, node)
		node = node.Parent
	}
	// Reverse to get root-to-node order
	slices.Reverse(path)
	return path
}
