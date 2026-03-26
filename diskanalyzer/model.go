package diskanalyzer

import "time"

// FileEntry represents a single file on disk.
type FileEntry struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	Ext     string    `json:"ext"` // lowercased, e.g. ".mp3"
}

// DirNode represents a directory in the tree.
// TotalSize is the recursive sum of all descendant file sizes.
type DirNode struct {
	Name      string     `json:"name"`
	Path      string     `json:"path"`
	TotalSize int64      `json:"total_size"`
	Files     []FileEntry `json:"files"`
	Children  []*DirNode  `json:"children"`
	Parent    *DirNode    `json:"-"` // nil for root, excluded from JSON
}

// AnalysisResult is the complete output of a Walk call.
type AnalysisResult struct {
	Root          *DirNode         `json:"root"`
	AllFiles      []FileEntry      `json:"all_files"`       // flat, sorted by size desc
	TotalSize     int64            `json:"total_size"`
	FileCount     int              `json:"file_count"`
	TypeBreakdown map[string]int64 `json:"type_breakdown"` // ext → total bytes
	ScannedAt     time.Time        `json:"scanned_at"`
}

// NewAnalysisResult creates a new AnalysisResult with initialized fields.
func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		AllFiles:      make([]FileEntry, 0),
		TypeBreakdown: make(map[string]int64),
		ScannedAt:     time.Now(),
	}
}
