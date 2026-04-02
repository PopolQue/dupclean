package scanner

import (
	"context"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/corona10/goimagehash"
	"golang.org/x/image/webp"
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

// hashedPhoto holds a photo path with its computed perceptual hash
type hashedPhoto struct {
	path string
	hash *goimagehash.ImageHash
	info os.FileInfo
}

// Scan implements the Scanner interface for photo duplicate detection
//
// Context Support: The scan can be cancelled via opts.Context. When cancelled,
// the function returns partial results collected up to the cancellation point.
func (s *PhotoScanner) Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error) {
	startTime := time.Now()
	stats := ScanStats{}

	// Create default context if none provided
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if opts.SimilarityPct > 0 {
		s.SimilarityPct = opts.SimilarityPct
	}

	// Stage 1: Collect photo files
	photos := make([]string, 0)
	visitedInodes := make(map[uint64]bool)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Log access errors for visibility
			log.Printf("[PhotoScanner] Access error: %v", err)
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

		// Check if it's a supported photo format
		if !photoExtensions[ext] {
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

		photos = append(photos, path)
		stats.TotalScanned++
		return nil
	})
	if err != nil {
		return nil, stats, err
	}

	// Stage 2: Compute perceptual hashes
	hashed := make([]hashedPhoto, 0, len(photos))
	for _, path := range photos {
		hash, info, err := computePerceptualHash(path)
		if err != nil {
			// Log files that can't be decoded
			log.Printf("[PhotoScanner] Hash error for %s: %v", path, err)
			stats.Errors = append(stats.Errors, NewScanError(path, ErrFileHash, err))
			continue
		}
		hashed = append(hashed, hashedPhoto{
			path: path,
			hash: hash,
			info: info,
		})
	}

	// Stage 3: Group by similarity
	groups := s.groupBySimilarity(hashed)

	// Calculate stats
	for _, group := range groups {
		if len(group.Files) >= 2 {
			stats.TotalDupes += len(group.Files) - 1
			stats.WastedBytes += group.Files[0].Size * int64(len(group.Files)-1)
		}
	}

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

	// Decode image based on format
	ext := strings.ToLower(filepath.Ext(path))
	var img image.Image
	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(f)
	case ".png":
		img, err = png.Decode(f)
	case ".gif":
		img, err = gif.Decode(f)
	case ".webp":
		img, err = webp.Decode(f)
	default:
		// Try generic decoder
		img, _, err = image.Decode(f)
	}
	if err != nil {
		return nil, nil, err
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
	used := make([]bool, len(photos))
	groups := make([]DuplicateGroup, 0)

	// Calculate maximum Hamming distance for our similarity threshold
	// 64-bit hash, 90% similarity = max 6 bits different
	maxDistance := int((100 - s.SimilarityPct) * 64 / 100)
	if maxDistance < 1 {
		maxDistance = 1
	}

	for i, photo := range photos {
		if used[i] {
			continue
		}

		// Start a new group with this photo
		group := DuplicateGroup{
			Hash:       photo.hash.ToString(),
			Files:      []FileInfo{},
			Similarity: 100,
		}
		group.Files = append(group.Files, FileInfo{
			Path:    photo.path,
			Name:    filepath.Base(photo.path),
			Size:    photo.info.Size(),
			ModTime: photo.info.ModTime(),
			Hash:    photo.hash.ToString(),
		})
		used[i] = true

		// Find similar photos
		for j := i + 1; j < len(photos); j++ {
			if used[j] {
				continue
			}

			distance, _ := photo.hash.Distance(photos[j].hash)
			if distance <= maxDistance {
				group.Files = append(group.Files, FileInfo{
					Path:    photos[j].path,
					Name:    filepath.Base(photos[j].path),
					Size:    photos[j].info.Size(),
					ModTime: photos[j].info.ModTime(),
					Hash:    photos[j].hash.ToString(),
				})
				used[j] = true
			}
		}

		// Only add groups with 2+ photos
		if len(group.Files) >= 2 {
			groups = append(groups, group)
		}
	}

	return groups
}
