package handler

import "image"

// mockBarCoder always returns the image and error
type mockBarCoder struct {
	image.Image
	err        error
	lastFormat string
}

// QRCode returns the image and error set in the struct.
func (m *mockBarCoder) BarCode(format string, boardID string, width, height int) (image.Image, error) {
	m.lastFormat = format
	return m.Image, m.err
}
