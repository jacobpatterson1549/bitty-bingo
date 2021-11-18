package server

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

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

func TestHTTPSHandlerServeHTTP(t *testing.T) {
	for i, test := range httpsHandlerServeHTTPTests {
		w := httptest.NewRecorder()
		gameInfos := make([]gameInfo, len(test.gameInfos), cap(test.gameInfos))
		copy(gameInfos, test.gameInfos) // do not modify the test value
		wantGameInfos := make([]gameInfo, len(test.wantGameInfos))
		copy(wantGameInfos, test.wantGameInfos)
		h := handler{
			time:      test.time,
			gameInfos: gameInfos,
		}
		test.r.Header = test.header
		bingo.GameResetter.Seed(1257894001) // make board new board creation deterministic
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

func TestWithGzip(t *testing.T) {
	for i, test := range withGzipTests {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("abc123")) // same as wantBodyStart for non-gzip accepting
		})
		w := httptest.NewRecorder()
		r := httptest.NewRequest(methodGet, "/", nil)
		r.Header.Add(headerAcceptEncoding, test.acceptEncoding)
		withGzipHandler(h).ServeHTTP(w, r)
		contentEncoding := w.Header().Get(headerContentEncoding)
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

const (
	schemeHTTP                = "http"
	schemeHTTPS               = "https"
	host                      = "example.com"
	methodGet                 = "GET"
	methodPost                = "POST"
	headerContentType         = "Content-Type"
	headerLocation            = "Location"
	headerXContentTypeOptions = "X-Content-Type-Options"
	headerContentEncoding     = "Content-Encoding"
	headerAcceptEncoding      = "Accept-Encoding"
	headerContentDisposition  = "Content-Disposition"
	contentTypeTextHTML       = "text/html; charset=utf-8"
	contentTypeTextPlain      = "text/plain; charset=utf-8"
	contentTypeEncodedForm    = "application/x-www-form-urlencoded"
	contentTypeGzip           = "application/x-gzip"
	contentTypeSVG            = "application/svg"
	contentEncodingGzip       = "gzip"
	xContentTypeNoSniff       = "nosniff"
	acceptEncodingsCommon     = "gzip, deflate, br"
	board1257894001IDNumbers  = "DwgEDAoTGxAcGSopHygxNDIuOUBIQ0ZKAQIDBQYHCQsNDhESFBUWFxgaHR4gISIjJCUmJyssLS8wMzU2Nzg6Ozw9Pj9BQkRFR0lL"
	board1257894001ID         = "5zuTsMm6CTZAs7ad"
	urlPathGames              = "/"
	urlPathGame               = "/game"
	urlPathGameCheckBoard     = "/game/board/check"
	urlPathGameBoard          = "/game/board"
	urlPathgameCreate         = "/game/create"
	urlPathGameDrawNumber     = "/game/draw_number"
	urlPathGameBoards         = "/game/boards"
	urlPathHelp               = "/help"
	urlPathAbout              = "/about"
	qpGameID                  = "gameID"
	qpBoardID                 = "boardID"
	qpType                    = "type"
	qpBingo                   = "bingo"
	typeHasLine               = "HasLine"
	typeIsFilled              = "IsFilled"
)

var httpHandlerTests = []struct {
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

var httpsHandlerTests = []struct {
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

var httpsHandlerServeHTTPTests = []struct {
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
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
		},
	},
	{
		name:           "get game",
		r:              httptest.NewRequest(methodGet, urlPathGame+"?"+qpGameID+"=5-"+board1257894001IDNumbers, nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
		},
	},
	{
		name:           "get game (zero)",
		r:              httptest.NewRequest(methodGet, urlPathGame, nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
		},
	},
	{
		name:           "check board - HasLine",
		r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeHasLine, nil),
		wantStatusCode: 303,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
			headerLocation:    {urlPathGame + "?" + qpGameID + "=5-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID + "&" + qpBingo},
		},
	},
	{
		name:           "check board - IsFilled",
		r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=24-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeIsFilled, nil),
		wantStatusCode: 303,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
			headerLocation:    {urlPathGame + "?" + qpGameID + "=24-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID + "&" + qpBingo},
		},
	},
	{
		name:           "check board - IsFilled (false)",
		r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=1-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeIsFilled, nil),
		wantStatusCode: 303,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
			headerLocation:    {urlPathGame + "?" + qpGameID + "=1-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID + ""},
		},
	},
	{
		name:           "get new board",
		r:              httptest.NewRequest(methodGet, urlPathGameBoard, nil),
		wantStatusCode: 303,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
			headerLocation:    {urlPathGameBoard + "?" + qpBoardID + "=" + board1257894001ID},
		},
	},
	{
		name:           "get board by id",
		r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+board1257894001ID, nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
		},
	},
	{
		name:           "help",
		r:              httptest.NewRequest(methodGet, urlPathHelp, nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
		},
	},
	{
		name:           "about",
		r:              httptest.NewRequest(methodGet, urlPathAbout, nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
		},
	},
	{
		name:           "create game",
		gameInfos:      []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
		wantGameInfos:  []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
		r:              httptest.NewRequest(methodGet, urlPathGame, nil),
		wantStatusCode: 200,
		wantHeader: http.Header{
			headerContentType: {contentTypeTextHTML},
		},
	},
	{
		name:      "draw number",
		time:      func() string { return "the_past_a" },
		gameInfos: append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}), // use append on empty slice with capacity
		wantGameInfos: []gameInfo{{
			ID:          "9-" + board1257894001IDNumbers,
			ModTime:     "the_past_a",
			NumbersLeft: 66,
		}, {ID: "1"}, {ID: "2"}, {ID: "3"}},
		r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"=8-"+board1257894001IDNumbers)),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
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
		r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"=8-"+board1257894001IDNumbers)),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
		wantStatusCode: 303,
		wantHeader: http.Header{
			headerLocation: {urlPathGame + "?" + qpGameID + "=9-" + board1257894001IDNumbers},
		},
	},
	{
		name:           "draw number - do not change game infos if all numbers are drawn",
		gameInfos:      append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
		wantGameInfos:  append(make([]gameInfo, 0, 10), gameInfo{ID: "1"}, gameInfo{ID: "2"}, gameInfo{ID: "3"}),
		r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"=75-"+board1257894001IDNumbers)),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
		wantStatusCode: 304,
		wantHeader: http.Header{
			headerLocation: {urlPathGame + "?" + qpGameID + "=75-" + board1257894001IDNumbers},
		},
	},
	{
		name:           "create boards",
		r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=5")),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
		wantStatusCode: 200,
		wantHeader: http.Header{
			headerContentType:        {"application/zip"},
			headerContentDisposition: {"attachment; filename=bingo-boards.zip"},
		},
	},
	{
		name:           "get game - bad id",
		r:              httptest.NewRequest(methodGet, urlPathGame+"?"+qpGameID+"=BAD-ID", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "get new board - bad id",
		r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"=BAD-ID", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "check board - bad game id",
		r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=BAD-ID&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeHasLine, nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "check board - bad board id",
		r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"=BAD-ID&"+qpType+"="+typeHasLine, nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "check board - bad check type",
		r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"=UNKNOWN", nil),
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	// {
	// 	name:           "get - not found",
	// 	r:              httptest.NewRequest(methodGet, "/UNKNOWN", nil),
	// 	wantStatusCode: 404,
	// 	wantHeader: http.Header{
	// 		headerContentType:         {contentTypeTextPlain},
	// 		headerXContentTypeOptions: {xContentTypeNoSniff},
	// 	},
	// },
	{
		name:           "draw number - no form content type header (cannot parse game id)",
		time:           func() string { return "" },
		gameInfos:      []gameInfo{{}},
		wantGameInfos:  []gameInfo{{}},
		r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"=8-"+board1257894001IDNumbers)),
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "draw number - bad game id",
		r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"=BAD-ID")),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "create boards - no form content type header (missing number)",
		r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=5")),
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "create boards - missing number",
		r:              httptest.NewRequest(methodPost, urlPathGameBoards, nil),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "create boards - number too small",
		r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=0")),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	{
		name:           "create boards - number too large",
		r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=9999999")),
		header:         http.Header{headerContentType: {contentTypeEncodedForm}},
		wantStatusCode: 400,
		wantHeader: http.Header{
			headerContentType:         {contentTypeTextPlain},
			headerXContentTypeOptions: {xContentTypeNoSniff},
		},
	},
	// {
	// 	name:           "post - not found",
	// 	r:              httptest.NewRequest(methodPost, "/UNKNOWN", nil),
	// 	wantStatusCode: 404,
	// 	wantHeader: http.Header{
	// 		headerContentType:         {contentTypeTextPlain},
	// 		headerXContentTypeOptions: {xContentTypeNoSniff},
	// 	},
	// },
	// {
	// 	name:           "bad method",
	// 	r:              httptest.NewRequest("DELETE", "/", nil),
	// 	wantStatusCode: 405,
	// 	wantHeader: http.Header{
	// 		headerContentType:         {contentTypeTextPlain},
	// 		headerXContentTypeOptions: {xContentTypeNoSniff},
	// 	},
	// },
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
		acceptEncoding: acceptEncodingsCommon,
		wantGzip:       true,
		wantBodyStart:  "\x1f\x8b\x08", // magic number (1f8b) and compression method for deflate (08)
	},
}
