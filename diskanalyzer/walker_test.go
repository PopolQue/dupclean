package diskanalyzer

import (
	"testing"

	"github.com/PopolQue/dupclean/internal/fsutil"
)

// ... existing tests

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
