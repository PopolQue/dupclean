package scanner

import (
	"fmt"
)

// ScanErrorType represents the type of scan error encountered
type ScanErrorType string

const (
	ErrFileRead    ScanErrorType = "FILE_READ"
	ErrFileHash    ScanErrorType = "FILE_HASH"
	ErrFileAccess  ScanErrorType = "FILE_ACCESS"
	ErrInvalidPath ScanErrorType = "INVALID_PATH"
	ErrSymlink     ScanErrorType = "SYMLINK"
	ErrIO          ScanErrorType = "IO"
	ErrUnknown     ScanErrorType = "UNKNOWN"
)

// ScanError represents an error that occurred during scanning
type ScanError struct {
	Path    string
	Type    ScanErrorType
	Err     error
	Skipped bool
}

func (e *ScanError) Error() string {
	if e.Path == "" {
		return fmt.Sprintf("[%s] %v", e.Type, e.Err)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Type, e.Path, e.Err)
}

func (e *ScanError) Unwrap() error { return e.Err }

// NewScanError creates a new ScanError
func NewScanError(path string, errType ScanErrorType, err error) *ScanError {
	return &ScanError{Path: path, Type: errType, Err: err, Skipped: false}
}

// NewSkippedError creates a new ScanError marked as skipped
func NewSkippedError(path string, errType ScanErrorType, err error) *ScanError {
	return &ScanError{Path: path, Type: errType, Err: err, Skipped: true}
}
