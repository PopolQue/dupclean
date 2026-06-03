package scanner

import (
	"errors"
	"os"
	"testing"
)

// TestScanError_Error tests the Error method
func TestScanError_Error(t *testing.T) {
	err := &ScanError{
		Path: "/test/file.txt",
		Type: ErrFileRead,
		Err:  os.ErrPermission,
	}

	expected := "[FILE_READ] /test/file.txt: permission denied"
	if err.Error() != expected {
		t.Errorf("ScanError.Error() = %q, want %q", err.Error(), expected)
	}
}

// TestScanError_Error_NoPath tests Error when path is empty
func TestScanError_Error_NoPath(t *testing.T) {
	err := &ScanError{
		Path: "",
		Type: ErrFileHash,
		Err:  errors.New("hash failed"),
	}

	expected := "[FILE_HASH] hash failed"
	if err.Error() != expected {
		t.Errorf("ScanError.Error() without path = %q, want %q", err.Error(), expected)
	}
}

// TestScanError_Unwrap tests error unwrapping
func TestScanError_Unwrap(t *testing.T) {
	originalErr := os.ErrPermission
	err := &ScanError{
		Path: "/test/file.txt",
		Type: ErrFileAccess,
		Err:  originalErr,
	}

	unwrapped := err.Unwrap()
	if unwrapped != originalErr {
		t.Error("ScanError.Unwrap() should return the original error")
	}
}

// TestNewScanError tests the constructor
func TestNewScanError(t *testing.T) {
	err := NewScanError("/test/file.txt", ErrFileRead, os.ErrNotExist)

	if err.Path != "/test/file.txt" {
		t.Errorf("Path = %q, want %q", err.Path, "/test/file.txt")
	}
	if err.Type != ErrFileRead {
		t.Errorf("Type = %q, want %q", err.Type, ErrFileRead)
	}
	if err.Err != os.ErrNotExist {
		t.Error("Err should be the original error")
	}
	if err.Skipped {
		t.Error("Skipped should be false for NewScanError")
	}
}

// TestNewSkippedError tests the skipped error constructor
func TestNewSkippedError(t *testing.T) {
	err := NewSkippedError("/test/file.txt", ErrFileAccess, os.ErrPermission)

	if !err.Skipped {
		t.Error("Skipped should be true for NewSkippedError")
	}
}

// TestScanErrorType_Values tests that all error type constants are defined
func TestScanErrorType_Values(t *testing.T) {
	types := []ScanErrorType{
		ErrFileRead,
		ErrFileHash,
		ErrFileAccess,
		ErrInvalidPath,
		ErrSymlink,
		ErrIO,
		ErrUnknown,
	}

	for _, typ := range types {
		if typ == "" {
			t.Errorf("Error type should not be empty: %v", typ)
		}
	}
}

// TestScanError_IsCompatibleWithErrorsIs tests compatibility with errors.Is
func TestScanError_IsCompatibleWithErrorsIs(t *testing.T) {
	err := NewScanError("/test.txt", ErrFileRead, os.ErrPermission)

	// Test that errors.Is works with the wrapped error
	if !errors.Is(err, os.ErrPermission) {
		t.Error("errors.Is should work with wrapped error")
	}
}
