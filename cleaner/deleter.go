package cleaner

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

// DeleteOptions configures how deletion is performed.
type DeleteOptions struct {
	DryRun      bool // report what would be deleted, delete nothing
	Permanent   bool // delete permanently instead of moving to trash
	Concurrency int  // worker pool size
	OnProgress  func(deleted int, totalBytes int64, current string)
}

// DeleteResult holds the result of a delete operation.
type DeleteResult struct {
	Deleted    int   // number of files/dirs deleted
	FreedBytes int64 // total bytes freed
	Skipped    int   // files that were open or permission-denied
	Errors     []DeleteError
}

// DeleteError represents an error during deletion.
type DeleteError struct {
	Path    string
	Err     error
	Skipped bool // true if file was skipped (e.g., in use)
}

// Delete deletes the given entries using trash or permanent deletion.
func Delete(entries []EntryInfo, opts DeleteOptions) (*DeleteResult, error) {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 4
	}

	result := &DeleteResult{
		Errors: make([]DeleteError, 0),
	}

	if opts.DryRun {
		// Just calculate what would be deleted
		for _, e := range entries {
			result.Deleted++
			result.FreedBytes += e.Size
		}
		return result, nil
	}

	// Worker pool
	jobs := make(chan EntryInfo, len(entries))
	results := make(chan deleteResult, len(entries))

	var wg sync.WaitGroup
	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				deleted, freed, skipped, err := deleteEntry(entry, opts.Permanent)
				results <- deleteResult{
					Deleted:    deleted,
					FreedBytes: freed,
					Skipped:    skipped,
					Error:      err,
					Path:       entry.Path,
				}
			}
		}()
	}

	// Feed jobs
	for _, e := range entries {
		jobs <- e
	}
	close(jobs)

	// Collect results
	wg.Wait()
	close(results)

	var deletedCount atomic.Int32
	var freedBytes atomic.Int64

	for r := range results {
		if r.Error != nil {
			result.Errors = append(result.Errors, DeleteError{
				Path:    r.Path,
				Err:     r.Error,
				Skipped: r.Skipped,
			})
			if r.Skipped {
				result.Skipped++
			}
		} else {
			deletedCount.Add(int32(r.Deleted))
			freedBytes.Add(r.FreedBytes)
		}

		if opts.OnProgress != nil {
			opts.OnProgress(int(deletedCount.Load()), freedBytes.Load(), r.Path)
		}
	}

	result.Deleted = int(deletedCount.Load())
	result.FreedBytes = freedBytes.Load()

	return result, nil
}

type deleteResult struct {
	Deleted    int
	FreedBytes int64
	Skipped    bool
	Error      error
	Path       string
}

// deleteEntry deletes a single entry.
func deleteEntry(entry EntryInfo, permanent bool) (deleted int, freedBytes int64, skipped bool, err error) {
	// Safety check: never delete empty paths
	if entry.Path == "" {
		return 0, 0, false, fmt.Errorf("cannot delete empty path")
	}

	// Safety check: never delete root directory
	if entry.Path == "/" || entry.Path == `\` {
		return 0, 0, false, fmt.Errorf("cannot delete root directory")
	}

	if permanent {
		// Permanent deletion
		err = os.RemoveAll(entry.Path)
		if err != nil {
			// Check if file is in use
			if isFileInUse(err) {
				return 0, 0, true, nil // skipped, not an error
			}
			return 0, 0, false, err
		}
		return 1, entry.Size, false, nil
	}

	// Move to trash using existing utility
	err = moveToTrash(entry.Path)
	if err != nil {
		// Check if file is in use
		if isFileInUse(err) {
			return 0, 0, true, nil // skipped
		}
		return 0, 0, false, err
	}

	return 1, entry.Size, false, nil
}

// isFileInUse checks if an error indicates a file is in use.
func isFileInUse(err error) bool {
	if err == nil {
		return false
	}
	// Check for common "file in use" errors
	errStr := err.Error()
	return containsAny(errStr, []string{
		"busy",
		"in use",
		"sharing violation",
		"permission denied",
		"access is denied",
	})
}

func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// moveToTrash moves a file or directory to the trash/recycle bin.
func moveToTrash(path string) error {
	return SafeMoveToTrash(path)
}
