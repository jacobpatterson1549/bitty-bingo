package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		wantOk bool
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
			wantOk: true,
		},
	}
	for i, test := range tests {
		s, err := test.cfg.NewServer()
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
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

func TestServerShutdown(t *testing.T) {
	s := Server{
		httpsServer: new(http.Server),
		httpServer:  new(http.Server),
	}
	ctx := context.Background()
	if err := s.Shutdown(ctx); err != nil {
		t.Errorf("unwanted error shutting down server: %v", err)
	}
}

func TestHTTPHandler(t *testing.T) {
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

func TestHTTPSHandler(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{
			GameCount: 10,
			Time: func() string {
				return "time"
			},
		}
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "https://example.com/", nil)
		h, err := cfg.httpsHandler()
		if err != nil {
			t.Fatalf("creating handler: %v", err)
		}
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
	})
	t.Run("bad config", func(t *testing.T) {
		var cfg Config
		if _, err := cfg.httpsHandler(); err == nil {
			t.Errorf("wanted error for bad config")
		}
	})
}

func TestGetBoardWithQR(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that depends on external library")
	}
	cfg := Config{
		GameCount: 10,
		Time: func() string {
			return "time"
		},
	}
	h, err := cfg.httpsHandler()
	if err != nil {
		t.Fatalf("error creating handler: %v", err)
	}
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
