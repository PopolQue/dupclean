package scanner

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// detectDuplicatesEngine is the shared 4-stage duplicate detection logic
func detectDuplicatesEngine(root string, opts Options, filter func(path string, info fs.FileInfo) bool) ([]DuplicateGroup, ScanStats, error) {
	start := time.Now()
	stats := ScanStats{}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	// Stage 1: Collect files and group by size
	bySize := make(map[int64][]string)
	visitedInodes := make(map[uint64]bool)
	fileCount := 0

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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

		// Check ignored extensions
		ext := strings.ToLower(filepath.Ext(d.Name()))
		for _, ignoredExt := range opts.IgnoreExtensions {
			if ext == ignoredExt {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Apply minimum size filter
		if info.Size() < opts.MinSize {
			return nil
		}

		// Apply custom filter (audio mode, etc)
		if filter != nil && !filter(path, info) {
			return nil
		}

		// Skip hard links using inode tracking
		if inode, ok := getInode(path, info); ok {
			if visitedInodes[inode] {
				return nil
			}
			visitedInodes[inode] = true
		}

		bySize[info.Size()] = append(bySize[info.Size()], path)
		fileCount++
		stats.TotalScanned++

		if opts.OnProgress != nil && (fileCount%10 == 0 || fileCount == 1) {
			opts.OnProgress(ScanProgress{
				Phase:      "Collecting files",
				Percent:    0.1,
				FilesFound: fileCount,
			})
		}

		return nil
	})

	if err != nil && err != ctx.Err() {
		return nil, stats, err
	}

	// Stage 2: Partial hash (concurrent)
	var stage2Paths []string
	for _, paths := range bySize {
		if len(paths) >= 2 {
			stage2Paths = append(stage2Paths, paths...)
		}
	}
	partialHashGroups := runConcurrentHashStage(ctx, stage2Paths, concurrency, func(path string) (string, error) {
		return hashFilePartial(path, DefaultPartialHashSize)
	}, opts.OnProgress, "Hashing files (partial)", 0.1, 0.4, &stats)

	// Stage 3: Full hash (concurrent)
	var stage3Paths []string
	for _, paths := range partialHashGroups {
		if len(paths) >= 2 {
			stage3Paths = append(stage3Paths, paths...)
		}
	}
	fullHashGroupsRaw := runConcurrentHashStage(ctx, stage3Paths, concurrency, func(path string) (string, error) {
		h, _, err := hashFileFull(path)
		return h, err
	}, opts.OnProgress, "Hashing files (full)", 0.4, 0.7, &stats)

	// Convert raw full hash groups to FileInfo groups
	fullHashGroups := make(map[string][]FileInfo)
	for hash, paths := range fullHashGroupsRaw {
		if len(paths) < 2 {
			continue
		}
		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}
			fullHashGroups[hash] = append(fullHashGroups[hash], FileInfo{
				Path:    path,
				Name:    filepath.Base(path),
				Size:    info.Size(),
				ModTime: info.ModTime(),
				Hash:    hash,
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
				stats.Errors = append(stats.Errors, NewScanError(files[i].Path, ErrIO, err))
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

type hashJob struct {
	path string
}

type hashResult struct {
	path string
	hash string
	err  error
}

func runConcurrentHashStage(ctx context.Context, allPaths []string, concurrency int, hashFn func(string) (string, error), onProgress func(ScanProgress), phase string, startPct, endPct float64, stats *ScanStats) map[string][]string {
	if len(allPaths) == 0 {
		return make(map[string][]string)
	}

	jobs := make(chan hashJob, len(allPaths))
	results := make(chan hashResult, len(allPaths))

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}

				hash, err := hashFn(job.path)
				results <- hashResult{path: job.path, hash: hash, err: err}
			}
		}()
	}

	for _, p := range allPaths {
		jobs <- hashJob{path: p}
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	outputGroups := make(map[string][]string)
	count := 0
	for res := range results {
		if res.err != nil {
			stats.Errors = append(stats.Errors, NewScanError(res.path, ErrFileHash, res.err))
			continue
		}
		outputGroups[res.hash] = append(outputGroups[res.hash], res.path)
		count++

		if onProgress != nil && (count%10 == 0 || count == 1) {
			onProgress(ScanProgress{
				Phase:       phase,
				Percent:     startPct + (float64(count)/float64(len(allPaths)))*(endPct-startPct),
				FilesHashed: count,
			})
		}
	}

	return outputGroups
}
