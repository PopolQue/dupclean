package fsutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// MeasureResult holds the result of measuring a directory.
type MeasureResult struct {
	TotalSize int64
	FileCount int
	DirCount  int
	Entries   []EntryInfo
}

// EntryInfo describes a single file or directory.
type EntryInfo struct {
	Path    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// MeasureDir recursively measures a directory and returns size/file count.
func MeasureDir(root string, patterns []string, minAge time.Duration) (*MeasureResult, error) {
	result := &MeasureResult{
		Entries: make([]EntryInfo, 0),
	}

	cutoff := time.Now().Add(-minAge)

	// Map to track size of directories and whether they match the pattern
	dirSizes := make(map[string]int64)
	dirMatched := make(map[string]bool)

	// Single pass: collect files and sizes
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}

		// Skip root itself
		if path == root {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Skip items newer than minAge
		if minAge > 0 && info.ModTime().After(cutoff) {
			return nil
		}

		// Check patterns
		matched := false
		if len(patterns) == 0 {
			matched = true
		} else {
			for _, pattern := range patterns {
				if m, _ := filepath.Match(pattern, d.Name()); m {
					matched = true
					break
				}
			}
		}

		if d.IsDir() {
			result.DirCount++
			if matched {
				dirMatched[path] = true
			}
		} else if matched {
			result.FileCount++
			result.TotalSize += info.Size()
			result.Entries = append(result.Entries, EntryInfo{
				Path:    path,
				Size:    info.Size(),
				ModTime: info.ModTime(),
				IsDir:   false,
			})
		}

		// Always update parent directory sizes for rollup
		parent := filepath.Dir(path)
		for {
			dirSizes[parent] += info.Size()
			if parent == root || parent == "." || parent == filepath.Dir(parent) {
				break
			}
			parent = filepath.Dir(parent)
		}

		return nil
	})

	// Add directory entries that matched the patterns
	for path, matched := range dirMatched {
		if matched {
			result.Entries = append(result.Entries, EntryInfo{
				Path:    path,
				Size:    dirSizes[path],
				ModTime: time.Now(),
				IsDir:   true,
			})
		}
	}

	return result, err
}

// computeDirSize computes the recursive size of a directory.
func computeDirSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total, err
}

// MeasureFile measures a single file.
func MeasureFile(path string) (*EntryInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &EntryInfo{
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}
