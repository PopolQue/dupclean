package cleaner

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"dupclean/internal/fsutil"
)

// ScanOptions configures how scanning is performed.
type ScanOptions struct {
	// Context controls cancellation of the scan operation.
	// If nil, the scan will run to completion.
	// Use context.WithTimeout() or context.WithCancel() to limit scan duration.
	Context context.Context

	Concurrency  int            // worker pool size; 0 = runtime.NumCPU()
	MinAge       time.Duration  // skip files newer than this (default: 0 = all files)
	MaxSize      int64          // cap entries reported per target (0 = unlimited)
	SkipPatterns []string       // additional glob exclusions
	OnProgress   func(Progress) // progress callback
}

// Progress holds scanning progress information.
type Progress struct {
	Total   int    // total targets
	Done    int    // targets fully scanned
	Current string // label of target currently being scanned
}

// Scan scans all targets concurrently and returns results.
//
// Context Support: The scan can be cancelled via opts.Context. When cancelled,
// the function returns partial results collected up to the cancellation point.
func Scan(targets []*CleanTarget, opts ScanOptions) (*ScanResult, error) {
	// Create default context if none provided
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if opts.Concurrency <= 0 {
		opts.Concurrency = runtime.NumCPU()
	}

	result := &ScanResult{
		Targets:   targets,
		ScannedAt: time.Now(),
		Errors:    make([]ScanError, 0),
	}

	// Filter targets with non-existent paths
	validTargets := make([]*CleanTarget, 0)
	for _, t := range targets {
		hasValidPath := false
		for _, p := range t.Paths {
			if _, err := os.Stat(p); err == nil {
				hasValidPath = true
				break
			}
		}
		if hasValidPath {
			validTargets = append(validTargets, t)
		}
	}
	targets = validTargets

	// Progress tracking
	var done atomic.Int32
	var currentTarget atomic.Value

	// Worker pool
	jobs := make(chan *CleanTarget, len(targets))
	results := make(chan *CleanTarget, len(targets))

	var wg sync.WaitGroup
	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				// Check for cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}

				scanTarget(target, opts)
				done.Add(1)

				// Send progress BEFORE sending result to avoid blocking
				if opts.OnProgress != nil {
					current := currentTarget.Load()
					if current == nil {
						current = ""
					}
					opts.OnProgress(Progress{
						Total:   len(targets),
						Done:    int(done.Load()),
						Current: current.(string),
					})
				}

				results <- target
			}
		}()
	}

	// Feed jobs
	for _, t := range targets {
		currentTarget.Store(t.Label)
		jobs <- t
	}
	close(jobs)

	// Collect results
	wg.Wait()
	close(results)

	for t := range results {
		result.TotalSize += t.TotalSize
	}

	return result, nil
}

// scanTarget scans a single target and populates its fields.
func scanTarget(target *CleanTarget, opts ScanOptions) {
	target.TotalSize = 0
	target.FileCount = 0
	target.Entries = make([]EntryInfo, 0)

	for _, path := range target.Paths {
		if _, err := os.Stat(path); err != nil {
			// Log non-existent paths for visibility
			log.Printf("[Cleaner] Path does not exist: %s - %v", path, err)
			continue
		}

		// Use fsutil to measure
		measureResult, err := fsutil.MeasureDir(path, target.Patterns, opts.MinAge)
		if err != nil {
			// Log measurement errors
			log.Printf("[Cleaner] Measure error for %s: %v", path, err)
			continue
		}

		target.TotalSize += measureResult.TotalSize
		target.FileCount += measureResult.FileCount

		// Convert fsutil.EntryInfo to cleaner.EntryInfo
		for _, e := range measureResult.Entries {
			target.Entries = append(target.Entries, EntryInfo{
				Path:    e.Path,
				Size:    e.Size,
				ModTime: e.ModTime,
				IsDir:   e.IsDir,
			})
		}
	}

	target.ScannedAt = time.Now()
}

// FilterTargets filters targets by category and IDs.
func FilterTargets(targets []*CleanTarget, category string, ids []string, noDeveloper, noBrowser bool) []*CleanTarget {
	if len(ids) > 0 {
		// Filter by specific IDs
		filtered := make([]*CleanTarget, 0)
		idMap := make(map[string]bool)
		for _, id := range ids {
			idMap[id] = true
		}
		for _, t := range targets {
			if idMap[t.ID] {
				filtered = append(filtered, t)
			}
		}
		return filtered
	}

	filtered := make([]*CleanTarget, 0)
	for _, t := range targets {
		// Skip excluded categories
		if noDeveloper && t.Category == "Developer" {
			continue
		}
		if noBrowser && t.Category == "Browser" {
			continue
		}

		// Filter by category if specified
		if category != "" && t.Category != category {
			continue
		}

		filtered = append(filtered, t)
	}

	return filtered
}

// GetTargetByPath finds the target that contains a given path.
func GetTargetByPath(targets []*CleanTarget, path string) *CleanTarget {
	for _, t := range targets {
		for _, p := range t.Paths {
			if rel, err := filepath.Rel(p, path); err == nil && !strings.HasPrefix(rel, "..") {
				return t
			}
		}
	}
	return nil
}
