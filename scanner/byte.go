package scanner

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ByteScanner implements duplicate detection for all file types using SHA-256
type ByteScanner struct{}

// NewByteScanner creates a new ByteScanner instance
func NewByteScanner() *ByteScanner {
	return &ByteScanner{}
}

// Scan implements the Scanner interface for general file duplicate detection
func (s *ByteScanner) Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error) {
	start := time.Now()
	stats := ScanStats{}

	// Track visited inodes to avoid following symlinks and hard links
	visitedInodes := make(map[uint64]bool)

	// Stage 1: Collect files and group by size
	bySize := make(map[int64][]string)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log access errors for visibility
			log.Printf("[ByteScanner] Access error: %v", err)
			return nil // skip unreadable files
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
				return nil // Already processed this inode
			}
			visitedInodes[inode] = true
		}

		bySize[info.Size()] = append(bySize[info.Size()], path)
		stats.TotalScanned++
		return nil
	})
	if err != nil {
		return nil, stats, err
	}

	// Stage 2-4: Multi-pass duplicate detection
	return s.detectDuplicates(bySize, start, stats)
}

// detectDuplicates performs the multi-stage duplicate detection algorithm
func (s *ByteScanner) detectDuplicates(bySize map[int64][]string, start time.Time, stats ScanStats) ([]DuplicateGroup, ScanStats, error) {
	// Stage 2: Partial hash (first 8KB)
	partialHashGroups := make(map[string][]string)
	for _, paths := range bySize {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			partialHash, err := hashFilePartial(path, partialHashSize)
			if err != nil {
				log.Printf("[ByteScanner] Partial hash error for %s: %v", path, err)
				continue
			}
			partialHashGroups[partialHash] = append(partialHashGroups[partialHash], path)
		}
	}

	// Stage 3: Full hash
	fullHashGroups := make(map[string][]FileInfo)
	for _, paths := range partialHashGroups {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			fullHash, info, err := hashFileFull(path)
			if err != nil {
				log.Printf("[ByteScanner] Full hash error for %s: %v", path, err)
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

	// Stage 4: Byte-by-byte verification
	var groups []DuplicateGroup
	for hash, files := range fullHashGroups {
		if len(files) < 2 {
			continue
		}

		// Verify with byte comparison
		verifiedFiles := []FileInfo{files[0]}
		for i := 1; i < len(files); i++ {
			match, err := filesIdentical(files[0].Path, files[i].Path)
			if err != nil {
				continue
			}
			if match {
				verifiedFiles = append(verifiedFiles, files[i])
			}
		}

		if len(verifiedFiles) >= 2 {
			groups = append(groups, DuplicateGroup{
				Hash:       hash,
				Files:      verifiedFiles,
				Similarity: 100,
			})
			stats.TotalDupes += len(verifiedFiles) - 1
			stats.WastedBytes += verifiedFiles[0].Size * int64(len(verifiedFiles)-1)
		}
	}

	stats.ScanDuration = time.Since(start)
	return groups, stats, nil
}
