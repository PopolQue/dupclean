package scanner

import (
	"crypto/sha256"
	"fmt"
	"io"
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

// FileInfo holds metadata about a scanned file
type FileInfo struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
	Hash    string
}

// DuplicateGroup is a set of files with identical content
type DuplicateGroup struct {
	Hash  string
	Files []FileInfo
}

// ScanStats holds summary statistics from a scan
type ScanStats struct {
	TotalScanned int
	TotalDupes   int
	WastedBytes  int64
	ScanDuration time.Duration
}

// ScanProgress holds progress information
type ScanProgress struct {
	Phase       string
	Percent     float64
	FilesFound  int
	FilesHashed int
}

// FindDuplicates scans a folder and returns groups of duplicate files
func FindDuplicates(folder string, includeAll bool, onProgress func(ScanProgress), ignoreFolders []string, ignoreExtensions []string) ([]DuplicateGroup, ScanStats, error) {
	start := time.Now()
	stats := ScanStats{}

	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Scanning for files...", Percent: 0})
	}

	// First pass: collect files by size (quick pre-filter)
	bySize := make(map[int64][]string)

	err := filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil // skip hidden files
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}

			// Check ignored folders
			for _, ignored := range ignoreFolders {
				if path == ignored || strings.HasPrefix(path, ignored+string(filepath.Separator)) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check ignored extensions
		ext := strings.ToLower(filepath.Ext(info.Name()))
		for _, ignoredExt := range ignoreExtensions {
			if ext == ignoredExt {
				return nil
			}
		}
		if !includeAll && !audioExtensions[ext] {
			return nil
		}

		bySize[info.Size()] = append(bySize[info.Size()], path)
		stats.TotalScanned++
		return nil
	})
	if err != nil {
		return nil, stats, err
	}

	// Count potential duplicates
	potentialDupes := 0
	for _, paths := range bySize {
		if len(paths) >= 2 {
			potentialDupes += len(paths)
		}
	}

	if onProgress != nil {
		onProgress(ScanProgress{
			Phase:      fmt.Sprintf("Found %d files (%d potential duplicates)", stats.TotalScanned, potentialDupes),
			Percent:    0.1,
			FilesFound: stats.TotalScanned,
		})
	}

	// Second pass: hash only files that share a size (potential duplicates)
	byHash := make(map[string][]FileInfo)

	hashCount := 0
	totalToHash := potentialDupes

	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Computing file hashes...", Percent: 0.15})
	}

	for _, paths := range bySize {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			hash, info, err := hashFile(path)
			if err != nil {
				continue
			}
			byHash[hash] = append(byHash[hash], FileInfo{
				Path:    path,
				Name:    filepath.Base(path),
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Hash:    hash,
			})
			hashCount++
			if onProgress != nil && totalToHash > 0 {
				percent := 0.15 + (float64(hashCount)/float64(totalToHash))*0.8
				onProgress(ScanProgress{
					Phase:       fmt.Sprintf("Hashing files... (%d/%d)", hashCount, totalToHash),
					Percent:     percent,
					FilesFound:  stats.TotalScanned,
					FilesHashed: hashCount,
				})
			}
		}
	}

	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Comparing hashes...", Percent: 0.95})
	}

	// Collect groups with 2+ files
	var groups []DuplicateGroup
	for hash, files := range byHash {
		if len(files) < 2 {
			continue
		}
		groups = append(groups, DuplicateGroup{Hash: hash, Files: files})
		stats.TotalDupes += len(files) - 1
		// Wasted bytes = (count - 1) * filesize
		stats.WastedBytes += files[0].Size * int64(len(files)-1)
	}

	stats.ScanDuration = time.Since(start)

	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Scan complete", Percent: 1.0})
	}

	return groups, stats, nil
}

func hashFile(path string) (string, os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return "", nil, err
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), info, nil
}
