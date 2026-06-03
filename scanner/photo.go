package scanner

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/corona10/goimagehash"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Supported photo extensions
var photoExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".bmp":  true,
	".tiff": true,
	".tif":  true,
}

// PhotoScanner implements duplicate detection for photos using perceptual hashing
type PhotoScanner struct {
	// SimilarityPct is the minimum similarity percentage (0-100)
	// Default is 90 (meaning images with ≤ 6-bit Hamming distance on 64-bit hash)
	SimilarityPct int
}

// NewPhotoScanner creates a new PhotoScanner instance
func NewPhotoScanner() *PhotoScanner {
	return &PhotoScanner{
		SimilarityPct: 90, // Default: 90% similar
	}
}

// Scan implements the Scanner interface for photo duplicate detection
func (s *PhotoScanner) Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error) {
	startTime := time.Now()
	stats := ScanStats{}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if opts.SimilarityPct > 0 {
		s.SimilarityPct = opts.SimilarityPct
	}

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = runtime.NumCPU()
	}

	// Stage 1: Collect photo files
	photos := make([]string, 0)
	visitedInodes := make(map[uint64]bool)

	err := walkFs(ctx, root, opts, visitedInodes, &stats, func(path string, info fs.FileInfo) bool {
		ext := strings.ToLower(filepath.Ext(path))
		// Check if extension is explicitly ignored in opts
		for _, ignoredExt := range opts.IgnoreExtensions {
			if ext == ignoredExt {
				return false
			}
		}
		// Then check if it's a photo extension
		return photoExtensions[ext]
	}, func(path string, info fs.FileInfo) error {
		photos = append(photos, path)
		stats.TotalScanned++

		if opts.OnProgress != nil && (len(photos)%10 == 0 || len(photos) == 1) {
			opts.OnProgress(ScanProgress{
				Phase:      "Collecting photos",
				Percent:    0.1,
				FilesFound: len(photos),
			})
		}
		return nil
	})

	if err != nil && err != ctx.Err() {
		return nil, stats, err
	}

	// Stage 2: Compute perceptual hashes (concurrent)
	type photoHashJob struct {
		path string
	}
	type photoHashResult struct {
		path string
		hash *goimagehash.ImageHash
		info os.FileInfo
		err  error
	}

	jobs := make(chan photoHashJob, len(photos))
	results := make(chan photoHashResult, len(photos))

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
				hash, info, err := computePerceptualHash(job.path)
				results <- photoHashResult{path: job.path, hash: hash, info: info, err: err}
			}
		}()
	}

	for _, p := range photos {
		jobs <- photoHashJob{path: p}
	}
	close(jobs)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in scanner photo collector: %v\n", r)
			}
		}()
		wg.Wait()
		close(results)
	}()

	hashed := make([]hashedPhoto, 0, len(photos))
	count := 0
	for res := range results {
		if res.err != nil {
			stats.Errors = append(stats.Errors, NewScanError(res.path, ErrFileHash, res.err))
			continue
		}
		hashed = append(hashed, hashedPhoto{
			path: res.path,
			hash: res.hash,
			info: res.info,
		})
		count++

		if opts.OnProgress != nil && (count%10 == 0 || count == 1) {
			opts.OnProgress(ScanProgress{
				Phase:       "Computing photo hashes",
				Percent:     0.1 + (float64(count)/float64(len(photos)))*0.7,
				FilesHashed: count,
			})
		}
	}

	// Stage 3: Group by similarity
	groups := s.groupBySimilarity(hashed)

	// Calculate stats
	for _, group := range groups {
		if len(group.Files) >= 2 {
			stats.TotalDupes += len(group.Files) - 1
			for i := 1; i < len(group.Files); i++ {
				stats.WastedBytes += group.Files[i].Size
			}
		}
	}

	stats.Mode = "photo"
	stats.ScanDuration = time.Since(startTime)
	return groups, stats, nil
}

// computePerceptualHash computes a perceptual hash for an image file
func computePerceptualHash(path string) (*goimagehash.ImageHash, os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}

	// Decode image (supports all formats registered via side-effect imports)
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, nil, err
	}

	// Check dimensions after decode to prevent OOM
	bounds := img.Bounds()
	if bounds.Dx() > 20000 || bounds.Dy() > 20000 {
		return nil, nil, fmt.Errorf("image dimensions too large: %dx%d", bounds.Dx(), bounds.Dy())
	}

	// Compute perceptual hash
	hash, err := goimagehash.PerceptionHash(img)
	if err != nil {
		return nil, nil, err
	}

	return hash, info, nil
}

// groupBySimilarity groups photos by perceptual hash similarity
func (s *PhotoScanner) groupBySimilarity(photos []hashedPhoto) []DuplicateGroup {
	if len(photos) == 0 {
		return nil
	}

	// Build BK-Tree for efficient similarity search
	tree := NewBKTree()
	for _, p := range photos {
		tree.Add(p)
	}

	used := make(map[string]bool)
	groups := make([]DuplicateGroup, 0)

	// Calculate maximum Hamming distance for our similarity threshold
	// 64-bit hash, 90% similarity = max 6 bits different
	maxDistance := int((100 - s.SimilarityPct) * 64 / 100)
	if maxDistance < 1 {
		maxDistance = 1
	}

	for _, photo := range photos {
		if used[photo.path] {
			continue
		}

		// Find similar photos using BK-Tree
		similar := tree.Search(photo.hash, maxDistance)

		// Filter out photos that are already used
		var groupPhotos []hashedPhoto
		for _, p := range similar {
			if !used[p.path] {
				groupPhotos = append(groupPhotos, p)
			}
		}

		// Only add groups with 2+ photos
		if len(groupPhotos) >= 2 {
			group := DuplicateGroup{
				Hash:       photo.hash.ToString(),
				Files:      make([]FileInfo, 0, len(groupPhotos)),
				Similarity: 100,
			}
			for _, p := range groupPhotos {
				group.Files = append(group.Files, FileInfo{
					Path:    p.path,
					Name:    filepath.Base(p.path),
					Size:    p.info.Size(),
					ModTime: p.info.ModTime(),
					Hash:    p.hash.ToString(),
				})
				used[p.path] = true
			}
			groups = append(groups, group)
		} else {
			// Mark as used even if no group is formed to avoid re-processing
			used[photo.path] = true
		}
	}

	return groups
}
