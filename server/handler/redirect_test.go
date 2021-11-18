package handler

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestRedirectHandler(t *testing.T) {
	wantStatusCode := 301
	for i, test := range redirectHandlerTests {
		w := httptest.NewRecorder()
		h := Redirect(test.httpsPort)
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

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
	host        = "example.com"
)

var redirectHandlerTests = []struct {
	name       string
	httpsPort  string
	r          *http.Request
	header     http.Header
	wantHeader http.Header
}{
	{
		name:      "default http port to default HTTP port",
		httpsPort: "443",
		r:         httptest.NewRequest(methodGet, schemeHTTP+"://"+host+"/", nil),
		wantHeader: http.Header{
			headerContentType: {contentTypeHTML},
			headerLocation:    {schemeHTTPS + "://" + host + "/"},
		},
	},
	{
		name:      "redirect to custom HTTPS port",
		httpsPort: "8000",
		r:         httptest.NewRequest(methodGet, schemeHTTP+"://"+host+":8001/", nil),
		wantHeader: http.Header{
			headerContentType: {contentTypeHTML},
			headerLocation:    {schemeHTTPS + "://" + host + ":8000/"},
		},
	},
}
