package scanner

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/corona10/goimagehash"
)

func BenchmarkHashFilePartial(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench_partial.txt")
	data := make([]byte, 1024*1024) // 1MB file
	if err := os.WriteFile(path, data, 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = hashFilePartial(path, DefaultPartialHashSize)
	}
}

func BenchmarkHashFileFull(b *testing.B) {
	sizes := []struct {
		name string
		size int64
	}{
		{"1MB", 1024 * 1024},
		{"10MB", 10 * 1024 * 1024},
	}

	for _, s := range sizes {
		b.Run(s.name, func(b *testing.B) {
			tmpDir := b.TempDir()
			path := filepath.Join(tmpDir, "bench_full.txt")
			data := make([]byte, s.size)
			if err := os.WriteFile(path, data, 0644); err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _, _ = hashFileFull(path)
			}
		})
	}
}

func BenchmarkFilesIdentical(b *testing.B) {
	tmpDir := b.TempDir()
	path1 := filepath.Join(tmpDir, "f1.txt")
	path2 := filepath.Join(tmpDir, "f2.txt")
	data := make([]byte, 1024*1024) // 1MB
	if err := os.WriteFile(path1, data, 0644); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(path2, data, 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = filesIdentical(path1, path2)
	}
}

func BenchmarkComputePerceptualHash(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "bench.png")

	// Create a simple image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 255, 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		b.Fatal(err)
	}
	f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = computePerceptualHash(path)
	}
}

func BenchmarkBKTreeSearch(b *testing.B) {
	counts := []int{100, 1000, 10000}

	for _, count := range counts {
		b.Run(fmt.Sprintf("Size_%d", count), func(b *testing.B) {
			tree := NewBKTree()
			var hashes []*goimagehash.ImageHash

			// Seed tree with slightly different hashes
			for i := 0; i < count; i++ {
				h := goimagehash.NewImageHash(uint64(i), goimagehash.PHash)
				hashes = append(hashes, h)
				tree.Add(hashedPhoto{
					path: fmt.Sprintf("/path/%d", i),
					hash: h,
				})
			}

			searchHash := hashes[len(hashes)/2]
			maxDistance := 6 // ~90% similarity

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tree.Search(searchHash, maxDistance)
			}
		})
	}
}
