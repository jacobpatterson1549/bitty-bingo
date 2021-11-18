package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithGzip(t *testing.T) {
	for i, test := range withGzipTests {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("abc123")) // same as wantBodyStart for non-gzip accepting
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(methodGet, "/", nil)
		r.Header.Add(headerAcceptEncoding, test.acceptEncoding)
		WithGzip(h).ServeHTTP(w, r)
		contentEncoding := w.Header().Get("Content-Encoding")
		gotGzip := contentEncoding == "gzip"
		gotMessage := w.Body.String()
		switch {
		case test.wantGzip != gotGzip:
			t.Errorf("test %v (%v): wanted gzip: %v, got %v", i, test.name, test.wantGzip, gotGzip)
		case !strings.HasPrefix(gotMessage, test.wantBodyStart):
			t.Errorf("test %v (%v): written message prefixes not equal:\nwanted: %x\ngot:    %x", i, test.name, test.wantBodyStart, gotMessage)
		}
	}
}

var withGzipTests = []struct {
	name           string
	acceptEncoding string
	wantGzip       bool
	wantBodyStart  string
	wantBody       string
}{
	{
		name:          "no accept encoding",
		wantBodyStart: "abc123",
	},
	{
		name:           "with gzip accept encoding",
		acceptEncoding: "gzip, deflate, br",
		wantGzip:       true,
		wantBodyStart:  "\x1f\x8b\x08", // magic number (1f8b) and compression method for deflate (08)
	},
}
