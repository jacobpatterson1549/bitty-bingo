package handler

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
			Request:        httptest.NewRequest(methodGet, "/", nil),
			wantStatusCode: 405,
		},
		{
			name:           "page not found",
			Mux:            Mux{methodGet: {"/b": okHandler}},
			Request:        httptest.NewRequest(methodGet, "/a", nil),
			wantStatusCode: 404,
		},
		{
			name:           "get to post endpoint",
			Mux:            Mux{methodGet: {"/": okHandler}},
			Request:        httptest.NewRequest(methodPost, "/", nil),
			wantStatusCode: 405,
		},
		{
			name:           "ok: single endpoint",
			Mux:            Mux{methodPost: {"/": okHandler}},
			Request:        httptest.NewRequest(methodPost, "/", nil),
			wantStatusCode: 200,
		},
		{
			name: "ok: multiple handlers",
			Mux: Mux{
				methodGet: {
					"/":       nil,
					"/help":   nil,
					"/page/a": nil,
					"/page/b": okHandler,
					"/page/c": nil,
				},
				methodPost: {
					"/create": nil,
				},
			},
			Request:        httptest.NewRequest(methodGet, "/page/b", nil),
			wantStatusCode: 200,
		},
	}
)
