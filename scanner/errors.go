package scanner

import (
	"fmt"
)

// ScanErrorType represents the type of scan error encountered
type ScanErrorType string

const (
	// ErrFileRead indicates a file could not be read
	ErrFileRead ScanErrorType = "FILE_READ"

	// ErrFileHash indicates a hashing error
	ErrFileHash ScanErrorType = "FILE_HASH"

	// ErrFileAccess indicates a file access permission error
	ErrFileAccess ScanErrorType = "FILE_ACCESS"

	// ErrInvalidPath indicates an invalid file path
	ErrInvalidPath ScanErrorType = "INVALID_PATH"

	// ErrSymlink indicates a symlink-related error
	ErrSymlink ScanErrorType = "SYMLINK"

	// ErrIO indicates a general I/O error
	ErrIO ScanErrorType = "IO"

	// ErrUnknown indicates an unknown error type
	ErrUnknown ScanErrorType = "UNKNOWN"
)

// ScanError represents an error that occurred during scanning
type ScanError struct {
	Path    string        // The file path that caused the error
	Type    ScanErrorType // The type of error
	Err     error         // The underlying error
	Skipped bool          // Whether the file was skipped due to this error
}

// Error implements the error interface for ScanError
func (e *ScanError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("[%s] %v", e.Type, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Type, e.Path, e.Err)
}

// Unwrap implements the errors.Unwrap interface for Go 1.13+
func (e *ScanError) Unwrap() error {
	return e.Err
}

// NewScanError creates a new ScanError with the given parameters
func NewScanError(path string, errType ScanErrorType, err error) *ScanError {
	return &ScanError{
		Path:    path,
		Type:    errType,
		Err:     err,
		Skipped: false,
	}
}

// NewSkippedError creates a new ScanError marked as skipped
func NewSkippedError(path string, errType ScanErrorType, err error) *ScanError {
	return &ScanError{
		Path:    path,
		Type:    errType,
		Err:     err,
		Skipped: true,
	}
}

// IsFileReadError checks if the error is a file read error
func (e *ScanError) IsFileReadError() bool {
	return e.Type == ErrFileRead
}

// IsHashError checks if the error is a hash error
func (e *ScanError) IsHashError() bool {
	return e.Type == ErrFileHash
}

// IsAccessError checks if the error is an access error
func (e *ScanError) IsAccessError() bool {
	return e.Type == ErrFileAccess
}

// IsSkipped checks if the file was skipped due to this error
func (e *ScanError) IsSkipped() bool {
	return e.Skipped
}

// ScanResult holds the result of a scan operation including any errors
type ScanResult struct {
	Groups []DuplicateGroup
	Stats  ScanStats
	Errors []*ScanError
}

// HasErrors returns true if the scan result contains any errors
func (r *ScanResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorCount returns the number of errors in the scan result
func (r *ScanResult) ErrorCount() int {
	return len(r.Errors)
}

// SkippedCount returns the number of files that were skipped
func (r *ScanResult) SkippedCount() int {
	count := 0
	for _, err := range r.Errors {
		if err.Skipped {
			count++
		}
	}
	return count
}
