package handler

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type (
	// gzipHandler wraps a handler to write it with gzipHandler encoding if the request accepts it.
	gzipHandler struct {
		http.Handler
	}
	// wrappedResponseWriter wraps response writing with another writer.
	wrappedResponseWriter struct {
		io.Writer
		http.ResponseWriter
	}
)

// WithGzip wraps the handler with a handler that writes responses using gzip compression when accepted.
func WithGzip(h http.Handler) http.Handler {
	return &gzipHandler{h}
}

// ServeHTTP writes all output with gzip encoding if the request allows it.
func (h *gzipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		h.Handler.ServeHTTP(w, r)
		return
	}
	gzw := gzip.NewWriter(w)
	defer gzw.Close()
	wrw := wrappedResponseWriter{
		Writer:         gzw,
		ResponseWriter: w,
	}
	wrw.Header().Set("Content-Encoding", "gzip")
	h.Handler.ServeHTTP(wrw, r)
}

// Write delegates the write to the wrapped writer.
func (wrw wrappedResponseWriter) Write(p []byte) (n int, err error) {
	return wrw.Writer.Write(p)
}
