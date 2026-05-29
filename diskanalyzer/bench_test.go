package diskanalyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkWalk(b *testing.B) {
	counts := []int{100, 1000}

	for _, count := range counts {
		b.Run(fmt.Sprintf("Files_%d", count), func(b *testing.B) {
			tmpDir := b.TempDir()

			// Create dummy structure
			for i := 0; i < count; i++ {
				dir := filepath.Join(tmpDir, fmt.Sprintf("dir_%d", i%10))
				_ = os.MkdirAll(dir, 0755)
				path := filepath.Join(dir, fmt.Sprintf("file_%d.txt", i))
				_ = os.WriteFile(path, []byte("some content"), 0644)
			}

			opts := DefaultOptions()
			opts.Concurrency = 4

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = Walk(tmpDir, opts)
			}
		})
	}
}
