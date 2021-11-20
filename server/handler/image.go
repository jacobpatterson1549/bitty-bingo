package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
	"github.com/jacobpatterson1549/bitty-bingo/server/handler/qr"
)

// qrCode creates an QR codeimage from the text.
var qrCode qrEncoder = qr.Image

type (
	// transparentImage wraps a gray16 image to make all points that are not black transparent.
	transparentImage struct {
		image.Image
	}
	// qrEncoder encodes text to a QR code with the specified size.
	qrEncoder func(text string, width, height int) (image.Image, error)
)

// newTransparentImage creates a transparentImage from the source image which must have gray 16 color model.
func newTransparentImage(m image.Image) (*transparentImage, error) {
	if cm := m.ColorModel(); cm != color.Gray16Model {
		return nil, fmt.Errorf("color model not gray16: %v", cm)
	}
	return &transparentImage{m}, nil
}

// At returns the color at a point on the image, Black is preserved from the source, everything else is transparent
func (m transparentImage) At(x, y int) color.Color {
	if c := m.Image.At(x, y); c == color.Black {
		return c
	}
	return color.Transparent
}

// ColorModel returns NRGBAModel.  This allows the transparentImage to encode transparent colors correctly to png.
func (transparentImage) ColorModel() color.Model {
	return color.NRGBAModel
}

// freeSpace converts the id of the board to a qr code, to a png image, and then encodes it with standard base64 encoding.
func freeSpace(b bingo.Board) (string, error) {
	id, err := b.ID()
	if err != nil {
		return "", fmt.Errorf("getting board id %v", err)
	}
	img, err := qrCode(id, 80, 80)
	if err != nil {
		return "", fmt.Errorf("creating qr image: %v", err)
	}
	var buf bytes.Buffer
	img = transparentImage{img}
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("converting barcode to png image: %v", err)
	}
	bytes := buf.Bytes()
	data := base64.StdEncoding.EncodeToString(bytes)
	return data, nil
}
