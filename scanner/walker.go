package scanner

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/PopolQue/dupclean/internal/fsutil"
)

// WalkFilter defines a custom filter function for the filesystem walk.
type WalkFilter func(path string, info fs.FileInfo) bool

// walkFs performs a standard filesystem walk with support for common filters.
func walkFs(ctx context.Context, root string, opts Options, visitedInodes map[uint64]bool, stats *ScanStats, filter WalkFilter, onFile func(string, fs.FileInfo) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			stats.Errors = append(stats.Errors, NewSkippedError(path, ErrFileAccess, err))
			return nil
		}

		// Security: Skip symlinks to prevent following malicious links
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}

		// Handle hidden files
		if !opts.IncludeHidden && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check ignored folders
		for _, ignored := range opts.IgnoreFolders {
			if path == ignored || strings.HasPrefix(path, ignored+string(filepath.Separator)) {
				return filepath.SkipDir
			}
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			stats.Errors = append(stats.Errors, NewSkippedError(path, ErrFileAccess, err))
			return nil
		}

		// Apply minimum size filter
		if info.Size() < opts.MinSize {
			return nil
		}

		// Apply custom filter
		if filter != nil && !filter(path, info) {
			return nil
		}

		// Skip hard links using inode tracking
		if inode, ok := fsutil.GetInode(path, info); ok {
			if visitedInodes[inode] {
				return nil
			}
			visitedInodes[inode] = true
		}

		return onFile(path, info)
	})
}
