package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

var audioExtensions = map[string]bool{
	".wav":  true,
	".aiff": true,
	".aif":  true,
	".mp3":  true,
	".flac": true,
	".ogg":  true,
	".m4a":  true,
	".aac":  true,
	".opus": true,
	".wma":  true,
}

// AudioScanner implements duplicate detection for audio files
type AudioScanner struct{}

// NewAudioScanner creates a new AudioScanner instance
func NewAudioScanner() *AudioScanner {
	return &AudioScanner{}
}

// Scan implements the Scanner interface for audio files
func (s *AudioScanner) Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error) {
	start := time.Now()
	stats := ScanStats{}

	// Stage 1: Collect files and group by size
	bySize := make(map[int64][]string)
	filesCollected := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
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

		// Audio mode: only scan audio files
		if !audioExtensions[ext] {
			return nil
		}

		// Apply minimum size filter
		if info.Size() < opts.MinSize {
			return nil
		}

		bySize[info.Size()] = append(bySize[info.Size()], path)
		stats.TotalScanned++
		filesCollected++

		// Report progress
		if opts.OnProgress != nil {
			opts.OnProgress(ScanProgress{
				Phase:      "Collecting files",
				Percent:    0.3, // Stage 1 is 30% of work
				FilesFound: filesCollected,
			})
		}

		return nil
	})
	if err != nil {
		return nil, stats, err
	}

	// Stage 2-4: Multi-pass duplicate detection
	return s.detectDuplicates(bySize, start, stats, opts.OnProgress)
}

// detectDuplicates performs the multi-stage duplicate detection algorithm
func (s *AudioScanner) detectDuplicates(bySize map[int64][]string, start time.Time, stats ScanStats, onProgress func(ScanProgress)) ([]DuplicateGroup, ScanStats, error) {
	// Count potential duplicates for progress
	potentialDupes := 0
	for _, paths := range bySize {
		if len(paths) >= 2 {
			potentialDupes += len(paths)
		}
	}

	// Stage 2: Partial hash (first 8KB)
	partialHashGroups := make(map[string][]string)
	hashCount := 0
	for _, paths := range bySize {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			partialHash, err := hashFilePartial(path, partialHashSize)
			if err != nil {
				continue
			}
			partialHashGroups[partialHash] = append(partialHashGroups[partialHash], path)
			hashCount++

			// Report progress
			if onProgress != nil {
				onProgress(ScanProgress{
					Phase:       "Hashing files (partial)",
					Percent:     0.3 + (float64(hashCount) / float64(potentialDupes) * 0.3),
					FilesHashed: hashCount,
				})
			}
		}
	}

	// Stage 3: Full hash
	fullHashGroups := make(map[string][]FileInfo)
	fullHashCount := 0
	for _, paths := range partialHashGroups {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			fullHash, info, err := hashFileFull(path)
			if err != nil {
				continue
			}
			fullHashGroups[fullHash] = append(fullHashGroups[fullHash], FileInfo{
				Path:    path,
				Name:    filepath.Base(path),
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Hash:    fullHash,
			})
			fullHashCount++

			// Report progress
			if onProgress != nil {
				onProgress(ScanProgress{
					Phase:       "Hashing files (full)",
					Percent:     0.6 + (float64(fullHashCount) / float64(potentialDupes) * 0.3),
					FilesHashed: fullHashCount,
				})
			}
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

// Legacy function for backwards compatibility
// Deprecated: Use AudioScanner.Scan() or ByteScanner.Scan() instead
func FindDuplicates(folder string, includeAll bool, onProgress func(ScanProgress), ignoreFolders []string, ignoreExtensions []string) ([]DuplicateGroup, ScanStats, error) {
	var scanner Scanner
	if includeAll {
		scanner = NewByteScanner()
	} else {
		scanner = NewAudioScanner()
	}
	
	opts := Options{
		IncludeHidden:    false,
		MinSize:          0,
		IgnoreFolders:    ignoreFolders,
		IgnoreExtensions: ignoreExtensions,
		OnProgress:       onProgress,
	}

	return scanner.Scan(folder, opts)
}
