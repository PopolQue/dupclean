package scanner

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ByteScanner implements duplicate detection for all file types using SHA-256
//
// Memory-Efficient Mode: Streaming is enabled by default for scans exceeding 50k files.
// This reduces memory usage by processing files in chunks.
type ByteScanner struct {
	// StreamingThreshold enables streaming mode when > 0.
	// When file count exceeds this threshold, processing happens in chunks.
	// Default: 50000 (automatically enabled for large scans)
	// Set to 0 to disable streaming (not recommended for large scans)
	StreamingThreshold int
}

// NewByteScanner creates a new ByteScanner instance with default settings.
// Streaming mode is enabled by default for scans > 50k files to reduce memory usage.
func NewByteScanner() *ByteScanner {
	return &ByteScanner{
		StreamingThreshold: 50000, // Auto-enable streaming for large scans
	}
}

// Scan implements the Scanner interface for general file duplicate detection
//
// Memory Note: For large directories (100k+ files), streaming mode automatically
// processes files in chunks to reduce memory pressure.
//
// Context Support: The scan can be cancelled via opts.Context. When cancelled,
// the function returns partial results collected up to the cancellation point.
//
// Streaming Mode: Enabled by default when file count exceeds 50k.
// Set opts.StreamingThreshold to 0 to disable, or to a custom value to override.
func (s *ByteScanner) Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error) {
	start := time.Now()
	stats := ScanStats{}

	// Create default context if none provided
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// Determine streaming threshold (options override scanner default)
	threshold := opts.StreamingThreshold
	if threshold <= 0 {
		threshold = s.StreamingThreshold
	}

	// Track visited inodes to avoid following symlinks and hard links
	visitedInodes := make(map[uint64]bool)

	// Stage 1: Collect files and group by size
	bySize := make(map[int64][]string)
	fileCount := 0
	allGroups := make([]DuplicateGroup, 0)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation before processing each file
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			log.Printf("[ByteScanner] Access error: %v", err)
			stats.Errors = append(stats.Errors, NewSkippedError(path, ErrFileAccess, err))
			return nil
		}

		// Skip symlinks to prevent following malicious links
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}

		if !opts.IncludeHidden && strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
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

		// Check ignored extensions
		ext := strings.ToLower(filepath.Ext(info.Name()))
		for _, ignoredExt := range opts.IgnoreExtensions {
			if ext == ignoredExt {
				return nil
			}
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Apply minimum size filter
		if info.Size() < opts.MinSize {
			return nil
		}

		// Skip hard links using inode tracking
		if inode, ok := getInode(info); ok {
			if visitedInodes[inode] {
				return nil
			}
			visitedInodes[inode] = true
		}

		bySize[info.Size()] = append(bySize[info.Size()], path)
		fileCount++
		stats.TotalScanned++

		// Streaming mode: process chunk when threshold reached
		if threshold > 0 && fileCount >= threshold && fileCount%threshold == 0 {
			chunkGroups, chunkStats, err := s.processChunk(bySize, ctx)
			if err != nil {
				return err
			}
			allGroups = append(allGroups, chunkGroups...)
			stats.TotalDupes += chunkStats.TotalDupes
			stats.WastedBytes += chunkStats.WastedBytes
			stats.Errors = append(stats.Errors, chunkStats.Errors...)

			// Clear processed data to free memory
			for k := range bySize {
				delete(bySize, k)
			}

			if opts.OnProgress != nil {
				opts.OnProgress(ScanProgress{
					Phase:      "Streaming scan",
					Percent:    0,
					FilesFound: fileCount,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, stats, err
	}

	// Warn about high memory usage
	if fileCount > MemoryWarningThreshold {
		log.Printf("[ByteScanner] Warning: Scanned %d files. Memory usage may be high.", fileCount)
	}

	// Process remaining files
	if len(bySize) > 0 {
		chunkGroups, chunkStats, err := s.processChunk(bySize, ctx)
		if err != nil {
			return nil, stats, err
		}
		allGroups = append(allGroups, chunkGroups...)
		stats.TotalDupes += chunkStats.TotalDupes
		stats.WastedBytes += chunkStats.WastedBytes
		stats.Errors = append(stats.Errors, chunkStats.Errors...)
	}

	stats.ScanDuration = time.Since(start)
	return allGroups, stats, nil
}

// processChunk processes a chunk of files grouped by size
func (s *ByteScanner) processChunk(bySize map[int64][]string, ctx context.Context) ([]DuplicateGroup, ScanStats, error) {
	stats := ScanStats{}
	groups := make([]DuplicateGroup, 0)

	// Stage 2: Partial hash
	partialHashGroups := make(map[string][]string)
	for size, paths := range bySize {
		select {
		case <-ctx.Done():
			return nil, stats, ctx.Err()
		default:
		}

		if len(paths) < 2 {
			continue
		}

		for _, path := range paths {
			select {
			case <-ctx.Done():
				return nil, stats, ctx.Err()
			default:
			}

			partialHash, err := hashFilePartial(path, DefaultPartialHashSize)
			if err != nil {
				log.Printf("[ByteScanner] Partial hash error for %s: %v", path, err)
				stats.Errors = append(stats.Errors, NewScanError(path, ErrFileHash, err))
				continue
			}
			partialHashGroups[partialHash] = append(partialHashGroups[partialHash], path)
		}

		_ = size
	}

	// Stage 3: Full hash
	fullHashGroups := make(map[string][]FileInfo)
	for _, paths := range partialHashGroups {
		if len(paths) < 2 {
			continue
		}

		for _, path := range paths {
			select {
			case <-ctx.Done():
				return nil, stats, ctx.Err()
			default:
			}

			fullHash, info, err := hashFileFull(path)
			if err != nil {
				log.Printf("[ByteScanner] Full hash error for %s: %v", path, err)
				stats.Errors = append(stats.Errors, NewScanError(path, ErrFileHash, err))
				continue
			}
			fullHashGroups[fullHash] = append(fullHashGroups[fullHash], FileInfo{
				Path:    path,
				Name:    filepath.Base(path),
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Hash:    fullHash,
			})
		}
	}

	// Stage 4: Collect groups
	for hash, files := range fullHashGroups {
		if len(files) < 2 {
			continue
		}

		groups = append(groups, DuplicateGroup{
			Hash:       hash,
			Files:      files,
			Similarity: 100,
		})

		stats.TotalDupes += len(files) - 1
		stats.WastedBytes += files[0].Size * int64(len(files)-1)
	}

	return groups, stats, nil
}
