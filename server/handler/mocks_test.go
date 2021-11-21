package handler

import "image"

// mockFreeSpacer always returns the image and error
type mockFreeSpacer struct {
	image.Image
	err error
}

// QRCode returns the image and error set in the struct.
func (f *mockFreeSpacer) QRCode(boardID string, width, height int) (image.Image, error) {
	return f.Image, f.err
}
