package scanner

import (
	"os"
	"path/filepath"
	"strings"
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
	return detectDuplicatesEngine(root, opts, "audio", func(path string, info os.FileInfo) bool {
		ext := strings.ToLower(filepath.Ext(info.Name()))
		return audioExtensions[ext]
	})
}

// FindDuplicates is a legacy function for backwards compatibility.
//
// Deprecated: Use AudioScanner.Scan() or ByteScanner.Scan() instead.
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
