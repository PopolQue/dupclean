package cleaner

import (
	"runtime"
	"testing"
)

// TestDeleteEntry_SafetyChecks tests that deleteEntry has proper safety checks
func TestDeleteEntry_SafetyChecks(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		permanent bool
		wantErr   bool
	}{
		{"empty path", "", true, true},
		{"empty path non-permanent", "", false, true},
		{"root unix", "/", true, true},
	}

	if runtime.GOOS == "windows" {
		tests = append(tests, struct {
			name      string
			path      string
			permanent bool
			wantErr   bool
		}{"root windows", `\`, true, true})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := EntryInfo{
				Path: tt.path,
				Size: 100,
			}

			deleted, freed, skipped, err := deleteEntry(entry, tt.permanent)

			if tt.wantErr && err == nil {
				t.Errorf("deleteEntry(%q) expected error, got nil", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("deleteEntry(%q) unexpected error: %v", tt.path, err)
			}
			if deleted != 0 {
				t.Errorf("deleteEntry(%q) deleted = %d, want 0", tt.path, deleted)
			}
			if freed != 0 {
				t.Errorf("deleteEntry(%q) freed = %d, want 0", tt.path, freed)
			}
			if skipped {
				t.Errorf("deleteEntry(%q) skipped = true, want false", tt.path)
			}
		})
	}
}

// TestDelete_SafetyEmptyPath tests that Delete handles empty paths safely
func TestDelete_SafetyEmptyPath(t *testing.T) {
	opts := DeleteOptions{
		Permanent:   true,
		Concurrency: 1,
	}

	// Empty path should error, not delete everything
	result, err := Delete([]EntryInfo{
		{Path: "", Size: 100},
	}, opts)

	// Error should be in the Errors slice
	if len(result.Errors) == 0 {
		t.Error("Delete with empty path should have an error in Errors slice")
	}
	if result.Deleted != 0 {
		t.Errorf("Delete with empty path should delete 0 files, got %d", result.Deleted)
	}
	if result.FreedBytes != 0 {
		t.Errorf("Delete with empty path should free 0 bytes, got %d", result.FreedBytes)
	}
	_ = err // err is nil, but error is in result.Errors
}

// TestDelete_SafetyRootPath tests that Delete handles root path safely
func TestDelete_SafetyRootPath(t *testing.T) {
	opts := DeleteOptions{
		Permanent:   true,
		Concurrency: 1,
	}

	// Root path should error, not delete everything
	result, err := Delete([]EntryInfo{
		{Path: "/", Size: 100},
	}, opts)

	// Error should be in the Errors slice
	if len(result.Errors) == 0 {
		t.Error("Delete with root path should have an error in Errors slice")
	}
	if result.Deleted != 0 {
		t.Errorf("Delete with root path should delete 0 files, got %d", result.Deleted)
	}
	if result.FreedBytes != 0 {
		t.Errorf("Delete with root path should free 0 bytes, got %d", result.FreedBytes)
	}
	_ = err // err is nil, but error is in result.Errors
}
