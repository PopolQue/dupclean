package diskanalyzer

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// WalkOptions configures how the filesystem walk is performed.
type WalkOptions struct {
	FollowSymlinks bool
	IncludeHidden  bool     // default false — skip dot files/folders
	MinSize        int64    // skip files smaller than this (bytes)
	MaxDepth       int      // 0 = unlimited
	ExcludePaths   []string // glob patterns, e.g. "node_modules", "*.git"
	Concurrency    int      // worker pool size; 0 = runtime.NumCPU()
}

// DefaultOptions returns WalkOptions with sensible defaults.
func DefaultOptions() WalkOptions {
	return WalkOptions{
		FollowSymlinks: false,
		IncludeHidden:  false,
		MinSize:        0,
		MaxDepth:       0,
		ExcludePaths:   []string{},
		Concurrency:    runtime.NumCPU(),
	}
}

// statResult holds the result of a stat operation.
type statResult struct {
	entry FileEntry
	err   error
}

// Walk traverses the filesystem root and returns an AnalysisResult.
// It runs in three passes:
// 1. Concurrent stat pass - collects file metadata
// 2. Tree-build pass - constructs DirNode hierarchy
// 3. Sort pass - orders by size for efficient queries
func Walk(root string, opts WalkOptions) (*AnalysisResult, []error, error) {
	if opts.Concurrency <= 0 {
		opts.Concurrency = runtime.NumCPU()
	}

	result := NewAnalysisResult()
	var walkErrors []error

	// Pass 1: Concurrent stat pass
	entries, errs, err := statPass(root, opts)
	if err != nil {
		return nil, nil, err
	}
	walkErrors = append(walkErrors, errs...)

	// Pass 2: Build tree and rollup sizes
	result.Root, err = buildTree(entries, root)
	if err != nil {
		return nil, walkErrors, fmt.Errorf("building tree: %w", err)
	}

	// Rollup sizes (post-order traversal)
	rollup(result.Root)

	// Build flat lists and aggregations
	result.AllFiles = entries
	sort.Slice(result.AllFiles, func(i, j int) bool {
		return result.AllFiles[i].Size > result.AllFiles[j].Size
	})

	result.TotalSize = result.Root.TotalSize
	result.FileCount = len(entries)

	// Build type breakdown
	for _, f := range entries {
		ext := f.Ext
		if ext == "" {
			ext = "(none)"
		}
		result.TypeBreakdown[ext] += f.Size
	}

	return result, walkErrors, nil
}

// statPass performs concurrent os.Stat on all files.
func statPass(root string, opts WalkOptions) ([]FileEntry, []error, error) {
	var entries []FileEntry
	var errors []error
	var mu sync.Mutex

	// Track visited inodes to avoid counting hard links multiple times
	visitedInodes := make(map[uint64]struct{})
	var inodeMu sync.Mutex

	// Channel for feeding paths to workers
	paths := make(chan string, opts.Concurrency*2)
	// Channel for collecting results
	results := make(chan statResult, opts.Concurrency*2)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range paths {
				info, err := os.Lstat(p)
				if err != nil {
					results <- statResult{err: err}
					continue
				}

				// Skip symlinks unless following them
				if info.Mode()&os.ModeSymlink != 0 && !opts.FollowSymlinks {
					continue
				}

				// Track inode to avoid hard link duplicates
				if inode, ok := getInode(info); ok {
					inodeMu.Lock()
					if _, seen := visitedInodes[inode]; seen {
						inodeMu.Unlock()
						continue
					}
					visitedInodes[inode] = struct{}{}
					inodeMu.Unlock()
				}

				// Skip hidden files
				if !opts.IncludeHidden && strings.HasPrefix(info.Name(), ".") {
					continue
				}

				// Skip files below min size
				if info.Size() < opts.MinSize {
					continue
				}

				// Check exclude patterns
				excluded := false
				for _, pattern := range opts.ExcludePaths {
					if matched, _ := filepath.Match(pattern, info.Name()); matched {
						excluded = true
						break
					}
				}
				if excluded {
					continue
				}

				results <- statResult{
					entry: FileEntry{
						Name:    info.Name(),
						Path:    p,
						Size:    info.Size(),
						ModTime: info.ModTime(),
						Ext:     strings.ToLower(filepath.Ext(p)),
					},
				}
			}
		}()
	}

	// Feeder goroutine - walks directory and feeds paths to workers
	go func() {
		depth := getDepth(root)
		if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Permission denied on directory - skip it
				return nil
			}

			// Check depth
			if opts.MaxDepth > 0 && getDepth(path)-depth > opts.MaxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip hidden directories
			if !opts.IncludeHidden && d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			// Skip symlinks to directories unless following
			if d.Type()&os.ModeSymlink != 0 && !opts.FollowSymlinks {
				return nil
			}

			if !d.IsDir() {
				paths <- path
			}
			return nil
		}); err != nil {
			log.Printf("Error walking directory: %v", err)
		}
		close(paths)
	}()

	// Collector goroutine - gathers results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect all results
	for r := range results {
		if r.err != nil {
			errors = append(errors, r.err)
		} else {
			mu.Lock()
			entries = append(entries, r.entry)
			mu.Unlock()
		}
	}

	return entries, errors, nil
}

// buildTree constructs a DirNode tree from a flat list of FileEntry.
func buildTree(entries []FileEntry, rootPath string) (*DirNode, error) {
	// Group files by directory
	dirFiles := make(map[string][]FileEntry)
	for _, f := range entries {
		dir := filepath.Dir(f.Path)
		dirFiles[dir] = append(dirFiles[dir], f)
	}

	// Get all unique directories
	dirs := make(map[string]struct{})
	for dir := range dirFiles {
		dirs[dir] = struct{}{}
		// Add parent directories
		for p := dir; p != rootPath && p != "."; p = filepath.Dir(p) {
			dirs[p] = struct{}{}
		}
	}

	// Create nodes
	nodes := make(map[string]*DirNode)
	for dir := range dirs {
		nodes[dir] = &DirNode{
			Name:  filepath.Base(dir),
			Path:  dir,
			Files: make([]FileEntry, 0),
		}
	}

	// Attach files to their directories
	for dir, files := range dirFiles {
		if node, ok := nodes[dir]; ok {
			node.Files = files
		}
	}

	// Build parent-child relationships
	root := nodes[rootPath]
	if root == nil {
		root = &DirNode{
			Name:  filepath.Base(rootPath),
			Path:  rootPath,
			Files: make([]FileEntry, 0),
		}
		nodes[rootPath] = root
	}

	for dir, node := range nodes {
		if dir == rootPath {
			continue
		}
		parentDir := filepath.Dir(dir)
		if parent, ok := nodes[parentDir]; ok {
			node.Parent = parent
			parent.Children = append(parent.Children, node)
		}
	}

	return root, nil
}

// rollup computes TotalSize for each node via post-order traversal.
func rollup(node *DirNode) {
	if node == nil {
		return
	}
	for _, child := range node.Children {
		rollup(child)
		node.TotalSize += child.TotalSize
	}
	for _, f := range node.Files {
		node.TotalSize += f.Size
	}
}

// getDepth returns the depth of a path relative to root.
func getDepth(path string) int {
	return strings.Count(filepath.Clean(path), string(filepath.Separator))
}
