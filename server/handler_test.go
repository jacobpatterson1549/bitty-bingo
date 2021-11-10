package server

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestHTTPHandler(t *testing.T) {
	for i, test := range httpHandlerTests {
		cfg := Config{
			HTTPSPort: test.httpsPort,
		}
		w := httptest.NewRecorder()
		h := cfg.httpHandler()
		h.ServeHTTP(w, test.r)
		wantStatusCode := 301
		test.wantHeader.Set("Content-Type", "text/html; charset=utf-8")
		gotStatusCode := w.Code
		gotHeader := w.Header()
		switch {
		case wantStatusCode != gotStatusCode:
			t.Errorf("test %v: status codes not equal: wanted %v, got %v", i, wantStatusCode, gotStatusCode)
		case !reflect.DeepEqual(test.wantHeader, gotHeader):
			t.Errorf("test %v: response headers not equal:\nwanted: %v\ngot:    %v", i, test.wantHeader, gotHeader)
		}
	}
}

func TestHTTPSHandler(t *testing.T) {
	t.Skip("TODO")
	// for i, test := range httpHandlerTests {
	// }
}

func TestWithGzip(t *testing.T) {
	for i, test := range withGzipTests {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("abc123")) // same as wantBodyStart for non-gzip accepting
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Add("Accept-Encoding", test.acceptEncoding)
		withGzip(h).ServeHTTP(w, r)
		contentEncoding := w.Header().Get("Content-Encoding")
		gotGzip := contentEncoding == "gzip"
		gotMessage := w.Body.String()
		switch {
		case test.wantGzip != gotGzip:
			t.Errorf("Test %v: wanted gzip: %v, got %v", i, test.wantGzip, gotGzip)
		case !strings.HasPrefix(gotMessage, test.wantBodyStart):
			t.Errorf("Test %v: written message prefixes not equal:\nwanted: %x\ngot:    %x", i, test.wantBodyStart, gotMessage)
		}
	}
}

var httpHandlerTests = []struct {
	httpsPort  string
	r          *http.Request
	wantHeader http.Header
}{
	{
		httpsPort: "443",
		r:         httptest.NewRequest("GET", "http://example.com", nil),
		wantHeader: http.Header{
			"Location": {"https://example.com"},
		},
	},
	{
		httpsPort: "8000",
		r:         httptest.NewRequest("GET", "http://example.com:8001", nil),
		wantHeader: http.Header{
			"Location": {"https://example.com:8000"},
		},
	},
}

// var httpsHandlerTests []struct {
// 	gameInfos []gameInfo
// }

var withGzipTests = []struct {
	acceptEncoding string
	wantGzip       bool
	wantBodyStart  string
}{
	{
		wantBodyStart: "abc123",
	},
	{
		acceptEncoding: "gzip, deflate, br",
		wantGzip:       true,
		wantBodyStart:  "\x1f\x8b\x08", // magic number (1f8b) and compression method for deflate (08)
	},
}
