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
			s.Config.Time = nil // functions are not comparable
			test.cfg.Time = nil // "
			if want, got := test.cfg, s.Config; !reflect.DeepEqual(want, got) {
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
	wantStatusCode := 301
	for i, test := range httpHandlerTests {
		w := httptest.NewRecorder()
		h := test.cfg.httpHandler()
		test.r.Header = test.header
		h.ServeHTTP(w, test.r)
		gotStatusCode := w.Code
		gotHeader := w.Header()
		switch {
		case wantStatusCode != gotStatusCode:
			t.Errorf("test %v (%v): HTTP response status codes not equal: wanted %v, got %v: %v", i, test.name, wantStatusCode, w.Code, w.Body.String())
		case !reflect.DeepEqual(test.wantHeader, gotHeader):
			t.Errorf("test %v (%v): HTTP response headers not equal:\nwanted: %v\ngot:    %v", i, test.name, test.wantHeader, gotHeader)
		}
	}
}

func TestHTTPSHandler(t *testing.T) {
	t.Run("valid configs", func(t *testing.T) {
		for i, test := range httpsHandlerTests {
			w := httptest.NewRecorder()
			h, err := test.cfg.httpsHandler()
			if err != nil {
				t.Errorf("test %v (%v): creating handler: %v", i, test.name, err)
				continue
			}
			test.r.Header = test.header
			h.ServeHTTP(w, test.r)
			gotStatusCode := w.Code
			gotHeader := w.Header()
			switch {
			case test.wantStatusCode != gotStatusCode:
				t.Errorf("test %v (%v): HTTPS response status codes not equal: wanted %v, got %v: %v", i, test.name, test.wantStatusCode, w.Code, w.Body.String())
			case !reflect.DeepEqual(test.wantHeader, gotHeader):
				t.Errorf("test %v (%v): HTTPS response headers not equal:\nwanted: %v\ngot:    %v", i, test.name, test.wantHeader, gotHeader)
			}
		}
	})
	t.Run("bad configs", func(t *testing.T) {
		tests := []struct {
			cfg  Config
			name string
		}{
			{
				name: "zero values [zero game count]",
			},
			{
				cfg: Config{
					GameCount: -9,
					Time: func() string {
						return "anything"
					},
				},
				name: "nonPositiveGameCount",
			},
			{
				cfg: Config{
					GameCount: 9,
				},
				name: "no time func",
			},
		}
		for i, test := range tests {
			if _, err := test.cfg.httpsHandler(); err == nil {
				t.Errorf("test %v (%v): wanted error for bad config: %#v", i, test.name, test.cfg)
			}
		}
	})
}

var (
	httpHandlerTests = []struct {
		name       string
		cfg        Config
		r          *http.Request
		header     http.Header
		wantHeader http.Header
	}{
		{
			name: "default http port to default HTTP port",
			cfg: Config{
				HTTPSPort: "443",
			},
			r: httptest.NewRequest(methodGet, schemeHTTP+"://"+host+"/", nil),
			wantHeader: http.Header{
				headerContentType: {contentTypeTextHTML},
				headerLocation:    {schemeHTTPS + "://" + host + "/"},
			},
		},
		{
			name: "redirect to custom HTTPS port",
			cfg: Config{
				HTTPSPort: "8000",
			},
			r: httptest.NewRequest(methodGet, schemeHTTP+"://"+host+":8001/", nil),
			wantHeader: http.Header{
				headerContentType: {contentTypeTextHTML},
				headerLocation:    {schemeHTTPS + "://" + host + ":8000/"},
			},
		},
		{
			name: "redirect with gzip",
			cfg: Config{
				HTTPSPort: "8000",
			},
			r: httptest.NewRequest(methodGet, schemeHTTP+"://"+host+":8001/", nil),
			header: http.Header{
				headerAcceptEncoding: {acceptEncodingsCommon},
			},
			wantHeader: http.Header{
				headerContentEncoding: {contentEncodingGzip},
				headerContentType:     {contentTypeTextHTML},
				headerLocation:        {schemeHTTPS + "://" + host + ":8000/"},
			},
		},
	}

	httpsHandlerTests = []struct {
		name           string
		cfg            Config
		r              *http.Request
		header         http.Header
		wantStatusCode int
		wantHeader     http.Header
	}{
		{
			name: "root with no accept encodings",
			cfg: Config{
				GameCount: 10,
				Time:      func() string { return "time" },
			},
			r:              httptest.NewRequest(methodGet, schemeHTTPS+"://"+host+"/", nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeTextHTML},
			},
		},
		{
			name: "root with gzip",
			cfg: Config{
				GameCount: 10,
				Time:      func() string { return "time" },
			},
			r: httptest.NewRequest(methodGet, schemeHTTPS+"://"+host+"/", nil),
			header: http.Header{
				headerAcceptEncoding: {acceptEncodingsCommon},
			},
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentEncoding: {contentEncodingGzip},
				headerContentType:     {contentTypeGzip},
			},
		},
		{
			name: "draw number",
			cfg: Config{
				GameCount: 10,
				Time:      func() string { return "then" },
			},
			r: httptest.NewRequest(methodPost, schemeHTTPS+"://"+host+""+urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"=8-"+board1257894001IDNumbers)),
			header: http.Header{
				headerContentType: {contentTypeEncodedForm},
			},
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {urlPathGame + "?" + qpGameID + "=9-" + board1257894001IDNumbers},
			},
		},
	}
)
