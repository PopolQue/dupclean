package scanner

import (
	"bytes"
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

const partialHashSize = 8 * 1024 // 8KB

// FindDuplicates scans a folder and returns groups of duplicate files
// Uses multi-stage detection: size → partial hash → full hash → byte comparison
func FindDuplicates(folder string, includeAll bool, onProgress func(ScanProgress), ignoreFolders []string, ignoreExtensions []string) ([]DuplicateGroup, ScanStats, error) {
	start := time.Now()
	stats := ScanStats{}

	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Scanning for files...", Percent: 0})
	}

	// Stage 1: Collect files and group by size
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

		// Check ignored folders
		for _, ignored := range ignoreFolders {
			if path == ignored || strings.HasPrefix(path, ignored+string(filepath.Separator)) {
				return filepath.SkipDir
			}
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
			Percent:    0.05,
			FilesFound: stats.TotalScanned,
		})
	}

	// Stage 2: Partial hash (first 8KB) for files with same size
	partialHashGroups := make(map[string][]string)
	hashCount := 0
	totalToHash := potentialDupes

	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Computing partial hashes (8KB)...", Percent: 0.1})
	}

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
			if onProgress != nil && totalToHash > 0 {
				percent := 0.1 + (float64(hashCount)/float64(totalToHash))*0.3
				onProgress(ScanProgress{
					Phase:       fmt.Sprintf("Partial hashing... (%d/%d)", hashCount, totalToHash),
					Percent:     percent,
					FilesFound:  stats.TotalScanned,
					FilesHashed: hashCount,
				})
			}
		}
	}

	// Stage 3: Full hash for files with matching partial hash
	fullHashGroups := make(map[string][]FileInfo)
	partialMatches := 0
	for _, paths := range partialHashGroups {
		if len(paths) >= 2 {
			partialMatches += len(paths)
		}
	}

	hashCount = 0
	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Computing full file hashes...", Percent: 0.4})
	}

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
			hashCount++
			if onProgress != nil && partialMatches > 0 {
				percent := 0.4 + (float64(hashCount)/float64(partialMatches))*0.4
				onProgress(ScanProgress{
					Phase:       fmt.Sprintf("Full hashing... (%d/%d)", hashCount, partialMatches),
					Percent:     percent,
					FilesFound:  stats.TotalScanned,
					FilesHashed: hashCount,
				})
			}
		}
	}

	// Stage 4: Byte-by-byte comparison for final verification
	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Verifying duplicates...", Percent: 0.8})
	}

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
				continue // skip verification failures
			}
			if match {
				verifiedFiles = append(verifiedFiles, files[i])
			}
		}

		if len(verifiedFiles) >= 2 {
			groups = append(groups, DuplicateGroup{Hash: hash, Files: verifiedFiles})
			stats.TotalDupes += len(verifiedFiles) - 1
			stats.WastedBytes += verifiedFiles[0].Size * int64(len(verifiedFiles)-1)
		}
	}

	stats.ScanDuration = time.Since(start)

	if onProgress != nil {
		onProgress(ScanProgress{Phase: "Scan complete", Percent: 1.0})
	}

	return groups, stats, nil
}

// hashFilePartial computes SHA256 of the first N bytes of a file
func hashFilePartial(path string, size int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	lr := io.LimitReader(f, size)
	if _, err := io.Copy(h, lr); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// hashFileFull computes SHA256 of the entire file
func hashFileFull(path string) (string, os.FileInfo, error) {
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

// filesIdentical performs byte-by-byte comparison of two files
func filesIdentical(path1, path2 string) (bool, error) {
	f1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	const bufSize = 32 * 1024 // 32KB buffers
	buf1 := make([]byte, bufSize)
	buf2 := make([]byte, bufSize)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		// Check for different lengths
		if n1 != n2 {
			return false, nil
		}

		// Check for different content
		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		// Check for EOF
		if err1 == io.EOF || err2 == io.EOF {
			return err1 == io.EOF && err2 == io.EOF, nil
		}

		// Check for errors
		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}
	}
}
