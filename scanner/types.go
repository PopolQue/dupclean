package scanner

import (
	"time"
)

// FileInfo holds metadata about a scanned file
type FileInfo struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
	Hash    string
}

// DuplicateGroup represents a set of files with identical or similar content
type DuplicateGroup struct {
	Hash     string
	Files    []FileInfo
	Similarity int // 100 for exact matches, <100 for similar (photos)
}

// ScanStats holds summary statistics from a scan
type ScanStats struct {
	TotalScanned int
	TotalDupes   int
	WastedBytes  int64
	ScanDuration time.Duration
}

// ScanProgress holds progress information during scanning
type ScanProgress struct {
	Phase       string
	Percent     float64
	FilesFound  int
	FilesHashed int
}

// Options configures how scanning is performed
type Options struct {
	// IncludeHidden scans hidden files and directories (default: false)
	IncludeHidden bool

	// MinSize filters files smaller than this (bytes)
	MinSize int64

	// Concurrency controls parallel hashing (default: runtime.NumCPU())
	Concurrency int

	// SimilarityPct is the minimum similarity percentage for photo mode (0-100)
	SimilarityPct int

	// IgnoreFolders is a list of folder paths to skip
	IgnoreFolders []string

	// IgnoreExtensions is a list of file extensions to skip
	IgnoreExtensions []string
}

// Scanner defines the interface for duplicate detection strategies
type Scanner interface {
	// Scan walks the directory tree and returns groups of duplicate files
	Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error)
}
