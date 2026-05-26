package scanner

// ByteScanner implements duplicate detection for all file types using SHA-256
//
// Memory-Efficient Mode: Streaming is enabled by default for scans exceeding 50k files.
// This reduces memory usage by processing files in chunks.
type ByteScanner struct {
	StreamingThreshold int
}

// Memory-Efficient Mode: Set opts.StreamingThreshold > 0 to enable streaming mode.
// In streaming mode, files are processed in chunks to reduce memory usage.

// NewByteScanner creates a new ByteScanner instance with default settings.
func NewByteScanner() *ByteScanner {
	return &ByteScanner{
		StreamingThreshold: 50000,
	}
}

// Scan implements the Scanner interface for general file duplicate detection
func (s *ByteScanner) Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error) {
	// ByteScanner uses the generic engine without additional filters
	return detectDuplicatesEngine(root, opts, nil)
}
