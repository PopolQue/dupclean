package scanner

// ByteScanner implements duplicate detection for all file types using SHA-256
type ByteScanner struct {
}

// NewByteScanner creates a new ByteScanner instance.
func NewByteScanner() *ByteScanner {
	return &ByteScanner{}
}

// Scan implements the Scanner interface for general file duplicate detection
func (s *ByteScanner) Scan(root string, opts Options) ([]DuplicateGroup, ScanStats, error) {
	// ByteScanner uses the generic engine without additional filters
	return detectDuplicatesEngine(root, opts, nil)
}
