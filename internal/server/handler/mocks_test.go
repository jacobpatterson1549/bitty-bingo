package handler

import "image"

// mockBarCoder always returns the image and error
type mockBarCoder struct {
	image.Image
	err error
}

// QRCode returns the image and error set in the struct.
func (f *mockBarCoder) BarCode(boardID string, width, height int) (image.Image, error) {
	return f.Image, f.err
}
