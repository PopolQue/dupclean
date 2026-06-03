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

// NewSkippedError creates a new skipped error.
func NewSkippedError(path string, errType ScanErrorType, err error) *ScanError {
	return &ScanError{
		Path:    path,
		Type:    errType,
		Err:     err,
		Skipped: true,
	}
}
