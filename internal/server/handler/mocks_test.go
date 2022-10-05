package handler

import (
	"bytes"
	"image"
	"net/http"
)

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

// mockResponseWriter is a very simple http.ResponseWriter.
// It can track if the header is written before all writes to the body.
type mockResponseWriter struct {
	header             http.Header
	writer             bytes.Buffer
	statusCode         int
	headerWrittenFirst bool
	writeCalled        bool
}

func (m *mockResponseWriter) Header() http.Header {
	m.headerWrittenFirst = !m.writeCalled
	return m.header
}

func (m *mockResponseWriter) Write(p []byte) (int, error) {
	m.writeCalled = true
	return m.writer.Write(p)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.headerWrittenFirst = !m.writeCalled
	m.statusCode = statusCode
}
