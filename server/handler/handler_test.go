package handler

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

func TestHandler(t *testing.T) {
	t.Run("valid configs", func(t *testing.T) {
		for i, test := range handlerTests {
			w := httptest.NewRecorder()
			h, err := Handler(test.gameCount, test.time)
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
	t.Run("bad parameters", func(t *testing.T) {
		tests := []struct {
			gameCount int
			time      func() string
			name      string
		}{
			{
				name: "zero values [zero game count]",
			},
			{
				gameCount: -9,
				time: func() string {
					return "anything"
				},
				name: "nonPositiveGameCount",
			},
			{
				gameCount: 9,
				name:      "no time func",
			},
		}
		for i, test := range tests {
			if _, err := Handler(test.gameCount, test.time); err == nil {
				t.Errorf("test %v (%v): wanted error for bad parameters", i, test.name)
			}
		}
	})
}

func TestHandlerServeHTTP(t *testing.T) {
	for i, test := range handlerServeHTTPTests {
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

const (
	methodGet                = "GET"
	methodPost               = "POST"
	headerContentType        = "Content-Type"
	headerLocation           = "Location"
	headerContentTypeOptions = "X-Content-Type-Options"
	headerAcceptEncoding     = "Accept-Encoding"
	headerContentDisposition = "Content-Disposition"
	contentTypeHTML          = "text/html; charset=utf-8"
	contentTypePlain         = "text/plain; charset=utf-8"
	contentTypeEncodedForm   = "application/x-www-form-urlencoded"
	ContentTypeNoSniff       = "nosniff"
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
	typeHasLine              = "HasLine"
	typeIsFilled             = "IsFilled"
)

var (
	handlerTests = []struct {
		name           string
		gameCount      int
		time           func() string
		r              *http.Request
		header         http.Header
		wantStatusCode int
		wantHeader     http.Header
	}{
		{
			name:           "root with no accept encodings",
			gameCount:      10,
			time:           func() string { return "time" },
			r:              httptest.NewRequest(methodGet, urlPathGames, nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
			},
		},
		{
			name:      "draw number",
			gameCount: 10,
			time:      func() string { return "then" },
			r:         httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader("gameID=8-"+board1257894001IDNumbers)),
			header: http.Header{
				headerContentType: {contentTypeEncodedForm},
			},
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {"/game?gameID=9-" + board1257894001IDNumbers},
			},
		},
	}
	handlerServeHTTPTests = []struct {
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
				headerContentType: {contentTypeHTML},
			},
		},
		{
			name:           "get game",
			r:              httptest.NewRequest(methodGet, urlPathGame+"?"+qpGameID+"=5-"+board1257894001IDNumbers, nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
			},
		},
		{
			name:           "get game (zero)",
			r:              httptest.NewRequest(methodGet, urlPathGame, nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
			},
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
				headerLocation:    {urlPathGame + "?" + qpGameID + "=1-" + board1257894001IDNumbers + "&" + qpBoardID + "=" + board1257894001ID + ""},
			},
		},
		{
			name:           "create board",
			r:              httptest.NewRequest(methodPost, urlPathGameBoard, nil),
			wantStatusCode: 303,
			wantHeader: http.Header{
				headerLocation: {urlPathGameBoard + "?" + qpBoardID + "=" + board1257894001ID},
			},
		},
		{
			name:           "get board by id",
			r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+board1257894001ID, nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
			},
		},
		{
			name:           "help",
			r:              httptest.NewRequest(methodGet, urlPathHelp, nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
			},
		},
		{
			name:           "about",
			r:              httptest.NewRequest(methodGet, urlPathAbout, nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
			},
		},
		{
			name:           "create game",
			gameInfos:      []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			wantGameInfos:  []gameInfo{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			r:              httptest.NewRequest(methodPost, urlPathGame, nil),
			wantStatusCode: 200,
			wantHeader: http.Header{
				headerContentType: {contentTypeHTML},
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
			r:              httptest.NewRequest(methodGet, urlPathGame+"?"+qpGameID+"="+badID, nil),
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "get new board - bad id",
			r:              httptest.NewRequest(methodGet, urlPathGameBoard+"?"+qpBoardID+"="+badID, nil),
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "check board - bad game id",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"="+badID+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+typeHasLine, nil),
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "check board - bad board id",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"="+badID+"&"+qpType+"="+typeHasLine, nil),
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "check board - bad check type",
			r:              httptest.NewRequest(methodGet, urlPathGameCheckBoard+"?"+qpGameID+"=5-"+board1257894001IDNumbers+"&"+qpBoardID+"="+board1257894001ID+"&"+qpType+"="+badID, nil),
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "get - not found",
			r:              httptest.NewRequest(methodGet, urlPathUnknown, nil),
			wantStatusCode: 404,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "draw number - no form content type header (cannot parse game id)",
			time:           func() string { return "" },
			gameInfos:      []gameInfo{{}},
			wantGameInfos:  []gameInfo{{}},
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"=8-"+board1257894001IDNumbers)),
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "draw number - bad game id",
			r:              httptest.NewRequest(methodPost, urlPathGameDrawNumber, strings.NewReader(""+qpGameID+"="+badID)),
			header:         http.Header{headerContentType: {contentTypeEncodedForm}},
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "create boards - no form content type header (missing number)",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=5")),
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "create boards - missing number",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, nil),
			header:         http.Header{headerContentType: {contentTypeEncodedForm}},
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "create boards - number too small",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=0")),
			header:         http.Header{headerContentType: {contentTypeEncodedForm}},
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "create boards - number too large",
			r:              httptest.NewRequest(methodPost, urlPathGameBoards, strings.NewReader("n=9999999")),
			header:         http.Header{headerContentType: {contentTypeEncodedForm}},
			wantStatusCode: 400,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "post - not found",
			r:              httptest.NewRequest(methodPost, urlPathUnknown, nil),
			wantStatusCode: 404,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
		{
			name:           "bad method",
			r:              httptest.NewRequest("DELETE", "/", nil),
			wantStatusCode: 405,
			wantHeader: http.Header{
				headerContentType:        {contentTypePlain},
				headerContentTypeOptions: {ContentTypeNoSniff},
			},
		},
	}
)
