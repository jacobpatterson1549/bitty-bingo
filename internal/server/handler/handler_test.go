package handler

import (
	"errors"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

func TestHandler(t *testing.T) {
	t.Run("valid configs", func(t *testing.T) {
		gameCount := 10
		timeF := func() string { return "any-time" }
		for i, test := range handlerTests {
			w := httptest.NewRecorder()
			h := Handler(gameCount, timeF, okMockBarCoder)
			test.r.Header = test.header
			h.ServeHTTP(w, test.r)
			gotStatusCode := w.Code
			gotHeader := w.Header()
			wantFaviconPrefix := "PHN2Z"
			switch {
			case test.wantStatusCode != gotStatusCode:
				t.Errorf("test %v (%v): HTTPS response status codes not equal: wanted %v, got %v: %v", i, test.name, test.wantStatusCode, w.Code, w.Body.String())
			case !reflect.DeepEqual(test.wantHeader, gotHeader):
				t.Errorf("test %v (%v): HTTPS response headers not equal:\nwanted: %v\ngot:    %v", i, test.name, test.wantHeader, gotHeader)
			case !strings.HasPrefix(h.(*handler).favicon, wantFaviconPrefix):
				t.Errorf("wanted favicon to be base64 encoded and start with %q [btoa('<svg')]:\n%v", wantFaviconPrefix, h.(*handler).favicon)
			}
		}
	})
	t.Run("zero configs", func(t *testing.T) {
		for i, test := range handlerTests {
			h := Handler(0, nil, nil)
			w := httptest.NewRecorder()
			test.r.Header = test.header
			h.ServeHTTP(w, test.r)
			if want, got := test.wantStatusCode, w.Code; want != got {
				t.Errorf("test %v (%v): wanted status code to be %v, got %v", i, test.name, want, got)
			}
		}
	})
}

func TestHandlerServeHTTP(t *testing.T) {
	for i, test := range handlerServeHTTPTests {
		w := httptest.NewRecorder()
		gameInfos := make([]gameInfo, len(test.gameInfos), cap(test.gameInfos))
		copy(gameInfos, test.gameInfos) // do not modify the test data, it is used in multiple tests
		wantGameInfos := make([]gameInfo, len(test.wantGameInfos))
		copy(wantGameInfos, test.wantGameInfos)
		h := handler{
			time:      test.time,
			gameInfos: gameInfos,
			BarCoder:  test.BarCoder,
		}
		test.r.Header = test.header
		bingo.GameResetter.Seed(1257894001) // make board creation deterministic
		h.ServeHTTP(w, test.r)
		switch {
		case w.Code != test.wantStatusCode:
			t.Errorf("test %v (%v): HTTPS response status codes not equal: wanted %v, got %v: %v", i, test.name, test.wantStatusCode, w.Code, w.Body.String())
		case !reflect.DeepEqual(test.wantHeader, w.Header()):
			t.Errorf("test %v (%v): HTTPS response headers not equal:\nwanted: %v\ngot:    %v", i, test.name, test.wantHeader, w.Header())
		case !reflect.DeepEqual(wantGameInfos, h.gameInfos):
			t.Errorf("test %v (%v): game infos not equal:\nwanted: %v\ngot:    %v", i, test.name, wantGameInfos, h.gameInfos)
		}
	}
}

func TestHandlerDrawNumberModifiesGames(t *testing.T) {
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(qpGameID+"=8-"+board1257894001IDNumbers))
	r1.Header = formContentTypeHeader
	h := handler{
		gameInfos: make([]gameInfo, 0, 1),
	}
	h.ServeHTTP(w1, r1)
	if want, got := 303, w1.Result().StatusCode; want != got {
		t.Fatalf("draw number status codes not equal: wanted %v, got %v", want, got)
	}
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest(methodGet, urlPathGames, nil)
	h.ServeHTTP(w2, r2)
	if want, got := "9-"+board1257894001IDNumbers, w2.Body.String(); !strings.Contains(got, want) {
		t.Errorf("wanted modified game history to be displayed on the games page (%v):\n%+v", want, w2)
	}
}

func TestHandlerBoardBarCode(t *testing.T) {
	r := image.Rect(0, 0, 256, 256)
	m := image.NewGray(r)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			m.Set(x, y, color.Gray{Y: uint8(x ^ y)})
		}
	}
	h := handler{
		BarCoder: &mockBarCoder{
			Image: m,
		},
	}
	got, err := h.boardBarCode(board1257894001ID, "barCodeFormat")
	switch {
	case err != nil:
		t.Errorf("unwanted error getting board bar code: %v", err)
	case !base64RE.MatchString(got):
		t.Errorf("wanted only base-64 standard encoding characters in bar code image, excluding right padding characters (=): (%v), got: %v", base64RE, got)
	}
}

func TestHandlerCreateBoards(t *testing.T) {
	w := httptest.NewRecorder()
	bc := mockBarCoder{
		Image: okMockBarCoder.Image,
	}
	h := handler{
		BarCoder: &bc,
	}
	bcFormat := "scribble"
	r := httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=5&barcodeFormat="+bcFormat))
	r.Header = formContentTypeHeader
	h.ServeHTTP(w, r)
	if want, got := bcFormat, bc.lastFormat; want != got {
		t.Errorf("bar code formats not equal: wanted %q, got %q", want, got)
	}
}

const (
	methodGet                = "GET"
	methodPost               = "POST"
	headerContentType        = "Content-Type"
	headerLocation           = "Location"
	contentTypeHTML          = "text/html; charset=utf-8"
	board1257894001IDNumbers = "DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL"
	board1257894001ID        = "5zuTsMm6CTZAs7ad"
	badID                    = "BAD-ID"
	urlPathGames             = "/"
	urlPathGame              = "/game"
	urlPathGameCheckBoard    = "/game/board/check"
	urlPathGameBoard         = "/game/board"
	urlPathGameDrawNumber    = "/game/draw_number"
	urlPathGameBoards        = "/game/boards"
	urlPathHelp              = "/help"
	urlPathAbout             = "/about"
	urlPathUnknown           = "/UNKNOWN"
	qpGameID                 = "gameID"
	qpBoardID                = "boardID"
	qpType                   = "type"
	qpBingo                  = "bingo"
	qpBarcodeFormat          = "barcodeFormat"
	typeHasLine              = "HasLine"
	typeIsFilled             = "IsFilled"
)

var (
	okMockBarCoder = &mockBarCoder{
		Image: image.NewGray16(image.Rect(0, 0, 1, 1)),
	}
	emptyImageMockMarCoder = &mockBarCoder{
		Image: image.NewGray16(image.Rect(0, 0, 0, 0)),
	}
	errMockBarCoder = &mockBarCoder{
		err: errors.New("mock error"),
	}
	base64RE              = regexp.MustCompile("^[a-zA-Z0-9+/]*={0,2}$")
	htmlContentTypeHeader = http.Header{
		headerContentType: {contentTypeHTML},
	}
	formContentTypeHeader = http.Header{
		headerContentType: {"application/x-www-form-urlencoded"},
	}
	errorHeader = http.Header{
		headerContentType:        {"text/plain; charset=utf-8"},
		"X-Content-Type-Options": {"nosniff"},
	}
	handlerTests = []struct {
		name           string
		r              *http.Request
		header         http.Header
		wantStatusCode int
		wantHeader     http.Header
	}{
		{
			name:           "root with no accept encodings",
			r:              httptest.NewRequest(methodGet, urlPathGames, nil),
			wantStatusCode: 200,
			wantHeader:     htmlContentTypeHeader,
		},
		{
			name:           "draw number",
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader("gameID=8-"+board1257894001IDNumbers)),
			header:         formContentTypeHeader,
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {urlPathGame + "?" + qpGameID + "=9-" + board1257894001IDNumbers},
			},
		},
		{
			name:           "view board",
			r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+board1257894001ID, nil),
			wantStatusCode: 200,
			wantHeader:     htmlContentTypeHeader,
		},
	}
	handlerServeHTTPTests = []struct {
		BarCoder
		name           string
		time           func() string
		gameInfos      []gameInfo
		wantGameInfos  []gameInfo
		r              *http.Request
		header         http.Header
		wantStatusCode int
		wantHeader     http.Header
	}{
		{
			name:           "get games list",
			r:              httptest.NewRequest(methodGet, urlPathGames, nil),
			wantStatusCode: 200,
			wantHeader:     htmlContentTypeHeader,
		},
		{
			name:           "get game",
			r:              httptest.NewRequest(methodGet, urlPathGame+"?"+qpGameID+"=5-"+board1257894001IDNumbers, nil),
			wantStatusCode: 200,
			wantHeader:     htmlContentTypeHeader,
		},
		{
			name:           "check board - HasLine",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeHasLine, nil),
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
				headerLocation:    {urlPathGame + "?" + qpGameID + "=5-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID + "&" + qpBingo},
			},
		},
		{
			name:           "check board - HasLine (false)",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=3-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeHasLine, nil),
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
				headerLocation:    {urlPathGame + "?" + qpGameID + "=3-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID},
			},
		},
		{
			name:           "check board - IsFilled",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=24-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeIsFilled, nil),
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
				headerLocation:    {urlPathGame + "?" + qpGameID + "=24-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID + "&" + qpBingo},
			},
		},
		{
			name:           "check board - IsFilled (false)",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=1-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeIsFilled, nil),
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
				headerLocation:    {urlPathGame + "?" + qpGameID + "=1-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID},
			},
		},
		{
			name:           "create board (preserves format)",
			r:              httptest.NewRequest(methodPost, urlPathGameBoard+"?"+qpBarcodeFormat+"=anything", nil),
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {urlPathGameBoard + "?" + qpBoardID + "=" + board1257894001ID + "&" + qpBarcodeFormat + "=anything"},
			},
		},
		{
			name:           "get board by id",
			r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+board1257894001ID, nil),
			BarCoder:       okMockBarCoder,
			wantStatusCode: 200,
			wantHeader:     htmlContentTypeHeader,
		},
		{
			name:           "help",
			r:              httptest.NewRequest(methodGet, urlPathHelp, nil),
			wantStatusCode: 200,
			wantHeader:     htmlContentTypeHeader,
		},
		{
			name:           "about",
			r:              httptest.NewRequest(methodGet, urlPathAbout, nil),
			wantStatusCode: 200,
			wantHeader:     htmlContentTypeHeader,
		},
		{
			name:           "create game",
			gameInfos:      []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			wantGameInfos:  []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			r:              httptest.NewRequest(methodPost, urlPathGame, nil),
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {urlPathGame + "?" + qpGameID + "=0"},
			},
		},
		{
			name:      "draw number",
			time:      func() string { return "the_past_a" },
			gameInfos: append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
			wantGameInfos: []gameInfo{{
				ID:          "9-" + board1257894001IDNumbers,
				ModTime:     "the_past_a",
				NumbersLeft: 66,
			}, {ID: "1"}, {ID: "2"}, {ID: "3"}},
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(qpGameID+"=8-"+board1257894001IDNumbers)),
			header:         formContentTypeHeader,
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {urlPathGame + "?" + qpGameID + "=9-" + board1257894001IDNumbers},
			},
		},
		{
			name:      "draw number (and discard last in history)",
			time:      func() string { return "the_past_b" },
			gameInfos: []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			wantGameInfos: []gameInfo{{
				ID:          "9-" + board1257894001IDNumbers,
				ModTime:     "the_past_b",
				NumbersLeft: 66,
			}, {ID: "1"}, {ID: "2"}},
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(qpGameID+"=8-"+board1257894001IDNumbers)),
			header:         formContentTypeHeader,
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {urlPathGame + "?" + qpGameID + "=9-" + board1257894001IDNumbers},
			},
		},
		{
			name:           "draw number - do not change game infos if all numbers are drawn",
			gameInfos:      append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
			wantGameInfos:  append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(qpGameID+"=75-"+board1257894001IDNumbers)),
			header:         formContentTypeHeader,
			wantStatusCode: 304,
			wantHeader:     http.Header{},
		},
		{
			name:           "create boards",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=5")),
			header:         formContentTypeHeader,
			BarCoder:       okMockBarCoder,
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType:     {"application/zip"},
				"Content-Disposition": {"attachment; filename=bingo-boards.zip"},
			},
		},
		{
			name:           "get game - bad id",
			r:              httptest.NewRequest(methodGet, urlPathGame+"?"+qpGameID+"="+badID, nil),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "get game - missing id",
			r:              httptest.NewRequest(methodGet, urlPathGame, nil),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "get new board - bad id",
			r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+badID, nil),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "check board - bad game id",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"="+badID+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeHasLine, nil),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "check board - bad board id",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"="+badID+"&"+qpType+"="+typeHasLine, nil),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "check board - bad check type",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+badID, nil),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "get board - BarCoder error",
			r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+board1257894001ID, nil),
			BarCoder:       errMockBarCoder,
			wantStatusCode: 500,
			wantHeader:     errorHeader,
		},
		{
			name:           "get board - BarCoder produces empty image",
			r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+board1257894001ID, nil),
			BarCoder:       emptyImageMockMarCoder,
			wantStatusCode: 500,
			wantHeader:     errorHeader,
		},
		{
			name:           "get - not found",
			r:              httptest.NewRequest(methodGet, urlPathUnknown, nil),
			wantStatusCode: 404,
			wantHeader:     errorHeader,
		},
		{
			name:           "draw number - no form content type header (cannot parse game id)",
			gameInfos:      []gameInfo{{}},
			wantGameInfos:  []gameInfo{{}},
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(qpGameID+"=8-"+board1257894001IDNumbers)),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "draw number - bad game id",
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(qpGameID+"="+badID)),
			header:         formContentTypeHeader,
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "create boards - no form content type header (missing number)",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=5")),
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "create boards - missing number",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, nil),
			header:         formContentTypeHeader,
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "create boards - number too small",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=0")),
			header:         formContentTypeHeader,
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "create boards - number too large",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=9999999")),
			header:         formContentTypeHeader,
			wantStatusCode: 400,
			wantHeader:     errorHeader,
		},
		{
			name:           "create boards - BarCoder error",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=1")),
			header:         formContentTypeHeader,
			BarCoder:       errMockBarCoder,
			wantStatusCode: 500,
			wantHeader:     errorHeader,
		},
		{
			name:           "post - not found",
			r:              httptest.NewRequest(methodPost, urlPathUnknown, nil),
			wantStatusCode: 404,
			wantHeader:     errorHeader,
		},
		{
			name:           "bad method",
			r:              httptest.NewRequest("DELETE", "/", nil),
			wantStatusCode: 405,
			wantHeader:     errorHeader,
		},
	}
)
