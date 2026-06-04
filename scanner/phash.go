package scanner

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"math/bits"
	"sort"
	"sync"
)

var (
	grayPool = sync.Pool{
		New: func() interface{} { return new([64][64]float64) },
	}
	dctPool = sync.Pool{
		New: func() interface{} { return new([64][64]float64) },
	}
	rgbaPool = sync.Pool{
		New: func() interface{} { return image.NewRGBA(image.Rect(0, 0, 64, 64)) },
	}
)

// PHash represents a 64-bit perceptual hash.
type PHash uint64

// Distance calculates the Hamming distance between two hashes.
func (h PHash) Distance(other PHash) (int, error) {
	return bits.OnesCount64(uint64(h ^ other)), nil
}

// String returns a string representation of the hash.
func (h PHash) String() string {
	return fmt.Sprintf("p:%016x", uint64(h))
}

// PerceptionHash computes a 64-bit perceptual hash (pHash) from an image.
// Algorithm: Resize to 64x64 -> Grayscale -> DCT -> Take top-left 8x8 -> Median -> Threshold
func PerceptionHash(img image.Image) (PHash, error) {
	// 1. Convert to RGBA for efficient processing
	var rgbaImg *image.RGBA
	if rgba, ok := img.(*image.RGBA); ok {
		rgbaImg = rgba
	} else {
		rgbaImg = rgbaPool.Get().(*image.RGBA)
		defer rgbaPool.Put(rgbaImg)
		draw.Draw(rgbaImg, rgbaImg.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	// 2. Resize to 64x64 (Nearest Neighbor)
	resized := rgbaPool.Get().(*image.RGBA)
	defer rgbaPool.Put(resized)
	resize64x64(rgbaImg, resized)

	// 3. Grayscale
	gray := grayPool.Get().(*[64][64]float64)
	defer grayPool.Put(gray)
	grayscale64x64(resized, gray)

	// 4. DCT
	dct := dctPool.Get().(*[64][64]float64)
	defer dctPool.Put(dct)
	dct64x64(gray, dct)

	// 5. Take top-left 8x8
	coeffs := [64]float64{} // Use array to avoid allocation
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			coeffs[y*8+x] = dct[y][x]
		}
	}

	// 6. Median
	sorted := coeffs
	sort.Float64Slice(sorted[:]).Sort()
	median := sorted[32]

	// 7. Threshold
	var hash uint64
	for i, c := range coeffs {
		if c > median {
			hash |= (1 << uint(i))
		}
	}

	return PHash(hash), nil
}

func resize64x64(img *image.RGBA, newImg *image.RGBA) {
	bounds := img.Bounds()
	dx := float64(bounds.Dx()) / 64.0
	dy := float64(bounds.Dy()) / 64.0

	for y := 0; y < 64; y++ {
		iy := int(float64(y)*dy) + bounds.Min.Y
		for x := 0; x < 64; x++ {
			ix := int(float64(x)*dx) + bounds.Min.X

			// Direct pixel access
			offset := (iy-bounds.Min.Y)*img.Stride + (ix-bounds.Min.X)*4
			newOffset := y*newImg.Stride + x*4
			copy(newImg.Pix[newOffset:newOffset+4], img.Pix[offset:offset+4])
		}
	}
}

func grayscale64x64(img *image.RGBA, gray *[64][64]float64) {
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			offset := y*img.Stride + x*4
			r := uint32(img.Pix[offset])
			g := uint32(img.Pix[offset+1])
			b := uint32(img.Pix[offset+2])
			// Convert to Luma (0.299R + 0.587G + 0.114B)
			// r, g, b are already 8-bit, so no need to shift
			gray[y][x] = 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
		}
	}
}

func dct64x64(gray *[64][64]float64, dct *[64][64]float64) {
	for u := 0; u < 8; u++ {
		for v := 0; v < 8; v++ {
			sum := 0.0
			for y := 0; y < 64; y++ {
				for x := 0; x < 64; x++ {
					sum += gray[y][x] * math.Cos((2.0*float64(x)+1.0)*float64(u)*math.Pi/128.0) * math.Cos((2.0*float64(y)+1.0)*float64(v)*math.Pi/128.0)
				}
			}
			dct[v][u] = sum
		}
	}
}
