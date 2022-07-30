package handler

import "image"

// mockBarcoder always returns the image and error.
type mockBarcoder struct {
	image.Image
	err        error
	lastFormat string
}

// Barcode returns the image and error set in the struct.
func (m *mockBarcoder) Barcode(format string, boardID string, width, height int) (image.Image, error) {
	m.lastFormat = format
	return m.Image, m.err
}
