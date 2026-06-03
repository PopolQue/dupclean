package cleaner

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
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
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic in deleter worker: %v\n", r)
				}
			}()
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

	var deletedCount atomic.Int64
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
			deletedCount.Add(int64(r.Deleted))
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

// protectedPaths contains directories that should never be permanently deleted.
var protectedPaths = []string{
	"/", "/bin", "/boot", "/dev", "/etc", "/home", "/lib", "/lib64", "/media", "/mnt", "/opt", "/proc", "/root", "/run", "/sbin", "/srv", "/sys", "/tmp", "/usr", "/var",
	"C:\\", "C:\\Windows", "C:\\Program Files", "C:\\Program Files (x86)", "C:\\Users",
}

// deleteEntry deletes a single entry.
func deleteEntry(entry EntryInfo, permanent bool) (deleted int, freedBytes int64, skipped bool, err error) {
	// Path validation for all operations
	if entry.Path == "" {
		return 0, 0, false, fmt.Errorf("cannot delete empty path")
	}

	abs, err := absPath(entry.Path)
	if err != nil {
		return 0, 0, false, fmt.Errorf("invalid path: %w", err)
	}

	if !permanent {
		// For trash operations, use the unified trash package which has built-in validation
		if err := moveToTrash(entry.Path); err != nil {
			// Check if file is in use
			if isFileInUse(err) {
				return 0, 0, true, nil // skipped, not an error
			}
			return 0, 0, false, err
		}
		return 1, entry.Size, false, nil
	}

	// Permanent deletion safety: ensure we're not deleting something we shouldn't
	// This is a "double-check" beyond the initial path validation.
	if err := verifyDeletionSafety(abs); err != nil {
		return 0, 0, false, err
	}

	// Permanent deletion
	err = osRemoveAll(entry.Path)

	if err != nil {
		// Check if file is in use
		if isFileInUse(err) {
			return 0, 0, true, nil // skipped, not an error
		}
		return 0, 0, false, err
	}
	return 1, entry.Size, false, nil
}

// verifyDeletionSafety performs a final check before permanent deletion.
func verifyDeletionSafety(path string) error {
	cleanPath := filepath.Clean(path)

	// Block roots
	if cleanPath == "/" || (goos == "windows" && len(cleanPath) <= 3 && strings.HasSuffix(cleanPath, ":\\")) {
		return fmt.Errorf("safety trigger: permanent deletion of root blocked: %s", path)
	}

	for _, protected := range protectedPaths {
		if cleanPath == filepath.Clean(protected) {
			return fmt.Errorf("safety trigger: permanent deletion of protected path blocked: %s", path)
		}
	}

	// Heuristic: Don't allow deleting the user's home directory itself
	home, err := userHomeDir()
	if err == nil {
		if cleanPath == filepath.Clean(home) {
			return fmt.Errorf("safety trigger: permanent deletion of home directory blocked: %s", path)
		}
	}

	return nil
}

// isFileInUse checks if an error indicates a file is in use or access is denied.
func isFileInUse(err error) bool {
	if err == nil {
		return false
	}

	// Robust check using standard errors
	if errors.Is(err, os.ErrPermission) {
		return true
	}

	// Check for "busy"
	if errors.Is(err, syscall.EBUSY) {
		return true
	}

	// Platform-specific checks via error codes
	if goos == "windows" {
		var errno syscall.Errno
		if errors.As(err, &errno) {
			switch errno {
			case 32: // ERROR_SHARING_VIOLATION
				return true
			case 5: // ERROR_ACCESS_DENIED
				return true
			}
		}
	}

	// Fallback to string matching for localized systems and wrapped errors
	errStr := strings.ToLower(err.Error())
	substrs := []string{
		"busy",
		"in use",
		"sharing violation",
		"access is denied",
		"permission denied",
		"operation not permitted",
	}
	for _, sub := range substrs {
		if strings.Contains(errStr, sub) {
			return true
		}
	}

	return false
}
