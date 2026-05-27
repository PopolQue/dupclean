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

	now := time.Now()
	cutoff := now.Add(-minAge)

	// To handle directory sizes correctly in a single pass without N^2 complexity,
	// we keep track of sizes for each directory level.
	dirSizes := make(map[string]int64)

	// First pass: collect all files and their sizes, and build directory size map
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

		// Skip files newer than minAge
		if minAge > 0 && info.ModTime().After(cutoff) {
			return nil
		}

		// Check patterns for the entry itself
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
		} else {
			if matched {
				result.FileCount++
				result.TotalSize += info.Size()
				result.Entries = append(result.Entries, EntryInfo{
					Path:    path,
					Size:    info.Size(),
					ModTime: info.ModTime(),
					IsDir:   false,
				})
			}

			// Always add file size to its parent directories for rollup
			parent := filepath.Dir(path)
			for {
				dirSizes[parent] += info.Size()
				if parent == root || parent == "." || parent == filepath.Dir(parent) {
					break
				}
				parent = filepath.Dir(parent)
			}
		}

		return nil
	})

	// Add directory entries that matched the patterns
	if len(patterns) > 0 {
		// Re-walk or filter to find matched directories and assign their rolled-up sizes
		// Actually, we can just iterate over the entries we've seen if we tracked them.
		// For simplicity and correctness with patterns, let's do a second pass over directories.
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || path == root || !d.IsDir() {
				return nil
			}

			matched := false
			for _, pattern := range patterns {
				if m, _ := filepath.Match(pattern, d.Name()); m {
					matched = true
					break
				}
			}

			if matched {
				size := dirSizes[path]
				// We don't add directory size to result.TotalSize here because its files
				// might have already been added if they matched.
				// Actually, the cleaner's logic usually treats a directory as a single unit
				// if it matches a pattern (like a cache folder).

				result.Entries = append(result.Entries, EntryInfo{
					Path:    path,
					Size:    size,
					ModTime: time.Now(), // ModTime for dir is less critical here
					IsDir:   true,
				})
			}
			return nil
		})
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
