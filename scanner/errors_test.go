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

// TestScanError_IsFileReadError tests error type checking
func TestScanError_IsFileReadError(t *testing.T) {
	err := NewScanError("/test/file.txt", ErrFileRead, os.ErrPermission)
	if !err.IsFileReadError() {
		t.Error("IsFileReadError() should return true")
	}

	err2 := NewScanError("/test/file.txt", ErrFileHash, os.ErrPermission)
	if err2.IsFileReadError() {
		t.Error("IsFileReadError() should return false for hash errors")
	}
}

// TestScanError_IsHashError tests hash error type checking
func TestScanError_IsHashError(t *testing.T) {
	err := NewScanError("/test/file.txt", ErrFileHash, os.ErrPermission)
	if !err.IsHashError() {
		t.Error("IsHashError() should return true")
	}

	err2 := NewScanError("/test/file.txt", ErrFileRead, os.ErrPermission)
	if err2.IsHashError() {
		t.Error("IsHashError() should return false for read errors")
	}
}

// TestScanError_IsAccessError tests access error type checking
func TestScanError_IsAccessError(t *testing.T) {
	err := NewScanError("/test/file.txt", ErrFileAccess, os.ErrPermission)
	if !err.IsAccessError() {
		t.Error("IsAccessError() should return true")
	}

	err2 := NewScanError("/test/file.txt", ErrFileRead, os.ErrPermission)
	if err2.IsAccessError() {
		t.Error("IsAccessError() should return false for read errors")
	}
}

// TestScanError_IsSkipped tests skipped status
func TestScanError_IsSkipped(t *testing.T) {
	err := NewSkippedError("/test/file.txt", ErrFileAccess, os.ErrPermission)
	if !err.IsSkipped() {
		t.Error("IsSkipped() should return true for skipped errors")
	}

	err2 := NewScanError("/test/file.txt", ErrFileRead, os.ErrPermission)
	if err2.IsSkipped() {
		t.Error("IsSkipped() should return false for regular errors")
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

// TestScanResult_HasErrors tests error checking
func TestScanResult_HasErrors(t *testing.T) {
	result := ScanResult{
		Groups: []DuplicateGroup{},
		Stats:  ScanStats{},
		Errors: []*ScanError{},
	}

	if result.HasErrors() {
		t.Error("HasErrors() should return false for empty errors")
	}

	result.Errors = append(result.Errors, NewScanError("/test.txt", ErrFileRead, os.ErrPermission))

	if !result.HasErrors() {
		t.Error("HasErrors() should return true when errors exist")
	}
}

// TestScanResult_ErrorCount tests error counting
func TestScanResult_ErrorCount(t *testing.T) {
	result := ScanResult{
		Errors: []*ScanError{
			NewScanError("/test1.txt", ErrFileRead, os.ErrPermission),
			NewSkippedError("/test2.txt", ErrFileAccess, os.ErrPermission),
			NewScanError("/test3.txt", ErrFileHash, os.ErrNotExist),
		},
	}

	if result.ErrorCount() != 3 {
		t.Errorf("ErrorCount() = %d, want 3", result.ErrorCount())
	}
}

// TestScanResult_SkippedCount tests skipped file counting
func TestScanResult_SkippedCount(t *testing.T) {
	result := ScanResult{
		Errors: []*ScanError{
			NewScanError("/test1.txt", ErrFileRead, os.ErrPermission),      // not skipped
			NewSkippedError("/test2.txt", ErrFileAccess, os.ErrPermission), // skipped
			NewScanError("/test3.txt", ErrFileHash, os.ErrNotExist),        // not skipped
			NewSkippedError("/test4.txt", ErrSymlink, os.ErrPermission),    // skipped
		},
	}

	if result.SkippedCount() != 2 {
		t.Errorf("SkippedCount() = %d, want 2", result.SkippedCount())
	}
}

// TestScanResult_Empty tests empty result
func TestScanResult_Empty(t *testing.T) {
	result := ScanResult{}

	if result.HasErrors() {
		t.Error("Empty result should not have errors")
	}
	if result.ErrorCount() != 0 {
		t.Errorf("Empty result ErrorCount() = %d, want 0", result.ErrorCount())
	}
	if result.SkippedCount() != 0 {
		t.Errorf("Empty result SkippedCount() = %d, want 0", result.SkippedCount())
	}
}

// TestScanError_ErrorTypes tests all error types work correctly
func TestScanError_ErrorTypes(t *testing.T) {
	tests := []struct {
		errType   ScanErrorType
		checkFunc func(*ScanError) bool
	}{
		{ErrFileRead, func(e *ScanError) bool { return e.IsFileReadError() }},
		{ErrFileHash, func(e *ScanError) bool { return e.IsHashError() }},
		{ErrFileAccess, func(e *ScanError) bool { return e.IsAccessError() }},
	}

	for _, tt := range tests {
		err := NewScanError("/test.txt", tt.errType, os.ErrPermission)
		if !tt.checkFunc(err) {
			t.Errorf("Error type %q check function should return true", tt.errType)
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
