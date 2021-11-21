package handler

import (
	"image"
	"image/color"
)

var (
	// pngColorModel is the png-friendly color model
	pngColorModel = color.NRGBAModel
	// transparentColor is the transparent color
	transparentColor = pngColorModel.Convert(color.Transparent)
)

// transparentImage wraps an image to make non-black pixels transparent.
type transparentImage struct {
	image.Image
	blackColor color.Color
}

// newTransparentImage creates a transparentImage from the source image which must have gray 16 color model.
func newTransparentImage(m image.Image) *transparentImage {
	return &transparentImage{
		Image:      m,
		blackColor: m.ColorModel().Convert(color.Black),
	}
}

// At returns the color at a point on the image, Black is preserved from the source, everything else is transparent.
func (m transparentImage) At(x, y int) color.Color {
	if c := m.Image.At(x, y); c == m.blackColor {
		return c
	}
	return transparentColor
}

// ColorModel returns NRGBAModel.  This allows the transparentImage to encode transparent colors correctly to png.
func (transparentImage) ColorModel() color.Model {
	return pngColorModel
}
