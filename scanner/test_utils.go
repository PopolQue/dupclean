package scanner

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

func getValidPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

var valid1x1PNG = getValidPNG()
