package diskanalyzer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PopolQue/dupclean/internal/fsutil"
)

func TestWalk(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a nested directory structure
	os.MkdirAll(filepath.Join(tmpDir, "sub1"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "f1.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "sub1", "f2.txt"), []byte("world"), 0644)

	opts := DefaultOptions()

	result, _, err := Walk(tmpDir, opts)
	if err != nil {
		t.Fatalf("Walk() error = %v", err)
	}

	if result == nil {
		t.Fatal("Walk() returned nil result")
	}

	if result.FileCount != 2 {
		t.Errorf("Walk() file count = %d, want 2", result.FileCount)
	}

	// Verify total size (5 + 5 = 10 bytes)
	if result.TotalSize != 10 {
		t.Errorf("Walk() total size = %d, want 10", result.TotalSize)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes  int64
		expect string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, test := range tests {
		result := fsutil.FormatBytes(test.bytes)
		if result != test.expect {
			t.Errorf("fsutil.FormatBytes(%d) = %s, want %s", test.bytes, result, test.expect)
		}
	}
}
