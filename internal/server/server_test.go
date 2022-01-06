package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "zero-value config [bad games count, time func]",
		},
		{
			name: "happy path",
			cfg: Config{
				GameCount: 100,
				Time: func() string {
					return "time"
				},
				HTTPPort:      "a",
				HTTPSPort:     "b",
				TLSCertFile:   "c",
				TLSKeyFile:    "d",
				HTTPSRedirect: true,
			},
		},
	}
	for i, test := range tests {
		s := test.cfg.NewServer()
		switch {
		case s.httpsServer == nil:
			t.Errorf("test %v (%v): HTTPS server not set", i, test.name)
		case s.httpServer == nil:
			t.Errorf("test %v (%v): HTTP server not set", i, test.name)
		default:
			s.config.Time = nil // functions are not comparable
			test.cfg.Time = nil // functions are not comparable
			if want, got := test.cfg, s.config; !reflect.DeepEqual(want, got) {
				t.Errorf("test %v (%v): config not copied to server:\nwanted: %v\ngot:    %v", i, test.name, want, got)
			}
		}
	}

}

func TestServerRunShutdown(t *testing.T) {
	tests := []struct {
		name string
		Config
	}{
		{
			name: "default values",
		},
		{
			name: "with HTTPS redirect",
			Config: Config{
				HTTPSRedirect: true,
			},
		},
	}
	for i, test := range tests {
		s := Server{
			httpsServer: new(http.Server),
			httpServer:  new(http.Server),
			config:      test.Config,
		}
		ctx := context.Background()
		s.Run()
		if err := s.Shutdown(ctx); err != nil {
			t.Errorf("test %v (%v): unwanted error shutting down server: %v", i, test.name, err)
		}
	}
}

func TestConfigHTTPHandler(t *testing.T) {
	cfg := Config{
		HTTPSPort: "8000",
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "http://example.com:8001/", nil)
	h := cfg.httpHandler()
	r.Header = http.Header{
		"Accept-Encoding": {"gzip, deflate, br"},
	}
	h.ServeHTTP(w, r)
	gotStatusCode := w.Code
	wantStatusCode := 301
	wantHeader := http.Header{
		"Content-Encoding": {"gzip"},
		"Content-Type":     {"text/html; charset=utf-8"},
		"Location":         {"https://example.com:8000/"},
	}
	gotHeader := w.Header()
	switch {
	case wantStatusCode != gotStatusCode:
		t.Errorf("HTTP response status codes not equal: wanted %v, got %v: %v", wantStatusCode, w.Code, w.Body.String())
	case !reflect.DeepEqual(wantHeader, gotHeader):
		t.Errorf("HTTP response headers not equal:\nwanted: %v\ngot:    %v", wantHeader, gotHeader)
	}
}

func TestConfigHTTPSHandler(t *testing.T) {
	var cfg Config
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "https://example.com/", nil)
	h := cfg.httpsHandler()
	r.Header = http.Header{
		"Accept-Encoding": {"gzip, deflate, br"},
	}
	h.ServeHTTP(w, r)
	wantStatusCode := 200
	gotStatusCode := w.Code
	wantHeader := http.Header{
		"Content-Encoding": {"gzip"},
		"Content-Type":     {"application/x-gzip"},
	}
	gotHeader := w.Header()
	switch {
	case wantStatusCode != gotStatusCode:
		t.Errorf("HTTPS response status codes not equal: wanted %v, got %v: %v", wantStatusCode, w.Code, w.Body.String())
	case !reflect.DeepEqual(wantHeader, gotHeader):
		t.Errorf("HTTPS response headers not equal:\nwanted: %v\ngot:    %v", wantHeader, gotHeader)
	}
}

func TestConfigHTTPSHandlerGetBoardWithBarCode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that depends on external library")
	}
	var cfg Config
	h := cfg.httpsHandler()
	boardID := "5zuTsMm6CTZAs7ad"
	r := httptest.NewRequest("GET", "/game/board?boardID="+boardID, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	got := w.Body.String()
	tests := []struct {
		name string
		want string
	}{
		{
			name: "The SVG image should be in document.",
			want: `<svg`,
		},
		{
			name: "The board's ID should be part of image.",
			want: boardID,
		},
		{
			name: "The QR codes seem to all start with this, the first 11 chars are from the png header.  This also checks the image width/height.",
			want: `width="80" height="80" href="data:image/png;base64,iVBORw0KGgo`,
		},
	}
	for i, test := range tests {
		if !strings.Contains(got, test.want) {
			t.Errorf("test %v (%v): response body did not contain %q:\n%v", i, test.name, test.want, got)
		}
	}
}

func TestFirstNonNilError(t *testing.T) {
	a := errors.New("a")
	b := errors.New("b")
	tests := []struct {
		name   string
		errors []error
		want   error
	}{
		{
			name: "no errors",
		},
		{
			name:   "all nil",
			errors: []error{nil, nil, nil},
			want:   nil,
		},
		{
			name:   "first set",
			errors: []error{a, nil},
			want:   a,
		},
		{
			name:   "second set",
			errors: []error{nil, b},
			want:   b,
		},
		{
			name:   "both set",
			errors: []error{a, b},
			want:   a,
		},
	}
	for i, test := range tests {
		if want, got := test.want, firstNonNilError(test.errors...); want != got {
			t.Errorf("test %v (%v): wanted %v, got %v", i, test.name, want, got)
		}
	}
}
