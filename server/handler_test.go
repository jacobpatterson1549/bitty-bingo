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
		gotStatusCode := w.Code
		gotHeader := w.Header()
		switch {
		case wantStatusCode != gotStatusCode:
			t.Errorf("test %v: response status codes not equal: wanted %v, got %v", i, wantStatusCode, gotStatusCode)
		case !reflect.DeepEqual(test.wantHeader, gotHeader):
			t.Errorf("test %v: response headers not equal:\nwanted: %v\ngot:    %v", i, test.wantHeader, gotHeader)
		}
	}
}

func TestHTTPSHandlerServeHTTP(t *testing.T) {
	for i, test := range httpsHandlerTests {
		w := httptest.NewRecorder()
		h := httpsHandler{
			time:      test.time,
			gameInfos: test.gameInfos,
		}
		test.r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		h.ServeHTTP(w, test.r)
		switch {
		case w.Code != test.wantStatusCode:
			t.Errorf("test %v: response status codes not equal: wanted %v, got %v: %v", i, test.wantStatusCode, w.Code, w.Body.String())
		case !reflect.DeepEqual(test.wantHeader, w.Header()):
			t.Errorf("test %v: response headers not equal:\nwanted: %v\ngot:    %v", i, test.wantHeader, w.Header())
		case !reflect.DeepEqual(test.wantGameInfos, h.gameInfos):
			t.Errorf("test %v: game infos not equal:\nwanted: %v\ngot:    %v", i, test.wantGameInfos, h.gameInfos)
		case len(test.wantBody) != 0 && test.wantBody != w.Body.String():
			t.Errorf("test %v: response bodies not equal:\nwanted: %v\ngot:    %v", i, test.wantBody, w.Body.String())
		}
	}
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
			"Content-Type": {"text/html; charset=utf-8"},
			"Location":     {"https://example.com"},
		},
	},
	{
		httpsPort: "8000",
		r:         httptest.NewRequest("GET", "http://example.com:8001", nil),
		wantHeader: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
			"Location":     {"https://example.com:8000"},
		},
	},
}

var httpsHandlerTests = []struct {
	time           func() string
	gameInfos      []gameInfo
	wantGameInfos  []gameInfo
	r              *http.Request
	wantStatusCode int
	wantHeader     http.Header
	wantBody       string // only checked if not empty
}{
	{ // get games list
		r:              httptest.NewRequest("GET", "/", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	},
	{ // get game
		r:              httptest.NewRequest("GET", "/game?gameID=5-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	},
	{ // get game (zero)
		r:              httptest.NewRequest("GET", "/game", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	},
	{ // check board - HasLine
		r:              httptest.NewRequest("GET", "/game/check_board?gameID=5-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL&boardID=5zuTsMm6CTZAs7ad&type=HasLine", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/plain; charset=utf-8"},
		},
		wantBody: "true",
	},
	{ // check board - IsFilled
		r:              httptest.NewRequest("GET", "/game/check_board?gameID=24-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL&boardID=5zuTsMm6CTZAs7ad&type=IsFilled", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/plain; charset=utf-8"},
		},
		wantBody: "true",
	},
	{ // check board - IsFilled (false)
		r:              httptest.NewRequest("GET", "/game/check_board?gameID=1-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL&boardID=5zuTsMm6CTZAs7ad&type=IsFilled", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/plain; charset=utf-8"},
		},
		wantBody: "false",
	},
	{ // help
		r:              httptest.NewRequest("GET", "/help", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	},
	{ // about
		r:              httptest.NewRequest("GET", "/about", nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			"Content-Type": {"text/html; charset=utf-8"},
		},
	},
	{ // create game
		gameInfos:      []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
		wantGameInfos:  []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
		r:              httptest.NewRequest("POST", "/game/create", nil),
		wantStatusCode: 303,
		wantHeader: http.Header{
			"Location": {"/game"},
		},
	},
	{ // draw number
		time:      func() string { return "the_past_a" },
		gameInfos: append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
		wantGameInfos: []gameInfo{{
			ID:          "9-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL",
			ModTime:     "the_past_a",
			NumbersLeft: 66,
		}, {ID: "1"}, {ID: "2"}, {ID: "3"}},
		r:              httptest.NewRequest("POST", "/game/draw_number", strings.NewReader("gameID=8-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL")),
		wantStatusCode: 303,
		wantHeader: http.Header{
			"Location": {"/game?gameID=9-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL"},
		},
	},
	{ // draw number (and discard last in history)
		time:      func() string { return "the_past_b" },
		gameInfos: append(make([]gameInfo, 0, 3), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
		wantGameInfos: []gameInfo{{
			ID:          "9-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL",
			ModTime:     "the_past_b",
			NumbersLeft: 66,
		}, {ID: "1"}, {ID: "2"}},
		r:              httptest.NewRequest("POST", "/game/draw_number", strings.NewReader("gameID=8-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL")),
		wantStatusCode: 303,
		wantHeader: http.Header{
			"Location": {"/game?gameID=9-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL"},
		},
	},
	{ // draw number - do not change game infos if all numbers are drawn
		gameInfos:      append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
		wantGameInfos:  append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
		r:              httptest.NewRequest("POST", "/game/draw_number", strings.NewReader("gameID=75-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL")),
		wantStatusCode: 304,
		wantHeader: http.Header{
			"Location": {"/game?gameID=75-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL"},
		},
	},
	{ // create boards
		r:              httptest.NewRequest("POST", "/game/boards", strings.NewReader("n=5")),
		wantStatusCode: 303,
		wantHeader: http.Header{
			"Content-Type":        {"application/zip"},
			"Content-Disposition": {"attachment; filename=bingo-boards.zip"},
			"Location":            {"/"},
		},
	},
	{ // get game - bad id
		r:              httptest.NewRequest("GET", "/game?gameID=BAD-ID", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // check board - bad game id
		r:              httptest.NewRequest("GET", "/game/check_board?gameID=BAD-ID&boardID=5zuTsMm6CTZAs7ad&type=HasLine", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // check board - bad board id
		r:              httptest.NewRequest("GET", "/game/check_board?gameID=5-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL&boardID=BAD-ID&type=HasLine", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // check board - bad check type
		r:              httptest.NewRequest("GET", "/game/check_board?gameID=5-DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL&boardID=5zuTsMm6CTZAs7ad&type=UNKNOWN", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // get - not found
		r:              httptest.NewRequest("GET", "/UNKNOWN", nil),
		wantStatusCode: 404,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // draw number - bad game id
		r:              httptest.NewRequest("POST", "/game/draw_number", strings.NewReader("gameID=BAD-ID")),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // create boards - missing number
		r:              httptest.NewRequest("POST", "/game/boards", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // create boards - number too small
		r:              httptest.NewRequest("POST", "/game/boards", strings.NewReader("n=0")),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // create boards - number too large
		r:              httptest.NewRequest("POST", "/game/boards", strings.NewReader("n=9999999")),
		wantStatusCode: 400,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // post - not found
		r:              httptest.NewRequest("POST", "/UNKNOWN", nil),
		wantStatusCode: 404,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
	{ // bad method
		r:              httptest.NewRequest("OPTIONS", "/", nil),
		wantStatusCode: 405,
		wantHeader: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
		},
	},
}

var withGzipTests = []struct {
	acceptEncoding string
	wantGzip       bool
	wantBodyStart  string
	wantBody       string
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
