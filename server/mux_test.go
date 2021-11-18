package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMuxServeHTTP(t *testing.T) {
	for i, test := range muxTests {
		w := httptest.NewRecorder()
		test.Mux.ServeHTTP(w, test.Request)
		if want, got := test.wantStatusCode, w.Code; want != got {
			t.Errorf("test %v (%v): status codes not equal: wanted %v, got %v", i, test.name, want, got)
		}
	}
}

const (
	methodGET                  = "GET"
	methodPOST                 = "POST"
	statusCodeOK               = 200
	statusCodeNotFound         = 404
	statusCodeMethodNotAllowed = 405
)

var (
	okHandler = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}
	muxTests = []struct {
		Mux
		*http.Request
		name           string
		wantStatusCode int
	}{
		{
			name:           "empty mux",
			Mux:            Mux{},
			Request:        httptest.NewRequest(methodGET, "/", nil),
			wantStatusCode: 405,
		},
		{
			name:           "page not found",
			Mux:            Mux{methodGET: {"/b": okHandler}},
			Request:        httptest.NewRequest(methodGET, "/a", nil),
			wantStatusCode: 404,
		},
		{
			name:           "get to post endpoint",
			Mux:            Mux{methodGET: {"/": okHandler}},
			Request:        httptest.NewRequest(methodPOST, "/", nil),
			wantStatusCode: 405,
		},
		{
			name:           "ok: single endpoint",
			Mux:            Mux{methodPOST: {"/": okHandler}},
			Request:        httptest.NewRequest(methodPOST, "/", nil),
			wantStatusCode: 200,
		},
		{
			name: "ok: multiple handlers",
			Mux: Mux{
				methodGET: {
					"/":       nil,
					"/help":   nil,
					"/page/a": nil,
					"/page/b": okHandler,
					"/page/c": nil,
				},
				methodPOST: {
					"/create": nil,
				},
			},
			Request:        httptest.NewRequest(methodGET, "/page/b", nil),
			wantStatusCode: 200,
		},
	}
)
