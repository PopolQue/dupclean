package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkScan(b *testing.B) {
	counts := []int{100, 1000}

	for _, count := range counts {
		b.Run(fmt.Sprintf("Files_%d", count), func(b *testing.B) {
			tmpDir := b.TempDir()

			// Create dummy structure for "system-temp" target
			tempDir := filepath.Join(tmpDir, "Temp")
			_ = os.MkdirAll(tempDir, 0755)
			for i := 0; i < count; i++ {
				path := filepath.Join(tempDir, fmt.Sprintf("file_%d.tmp", i))
				_ = os.WriteFile(path, []byte("some content"), 0644)
			}

			targets := []*CleanTarget{
				{
					ID:       "bench-temp",
					Paths:    []string{tempDir},
					Patterns: []string{"*.tmp"},
				},
			}

			opts := ScanOptions{
				Concurrency: 4,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = Scan(targets, opts)
			}
		})
	}
}
