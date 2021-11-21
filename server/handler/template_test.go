package handler

import (
	"bytes"
	"errors"
	"html/template"
	"image"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

var (
	okMockFreeSpacer = &mockFreeSpacer{
		Image: image.NewGray16(image.Rect(0, 0, 1, 1)),
	}
	errMockFreeSpacer = &mockFreeSpacer{
		err: errors.New("mock FreeSpacer error"),
	}
)

func TestNewTemplateGame(t *testing.T) {
	var g bingo.Game
	g.DrawNumber()
	g.DrawNumber()
	gameID := "game-464"
	boardID := "board-1797"
	hasBingo := true
	want := &game{
		Game:     g,
		GameID:   gameID,
		BoardID:  boardID,
		HasBingo: true,
	}
	got := newTemplateGame(g, gameID, boardID, hasBingo)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("template games not equal:\nwanted: %#v\ngot:    %#v", want, got)
	}
}

func TestNewTemplateBoard(t *testing.T) {
	b := bingo.NewBoard()
	boardID := "board-1341"
	want := &board{
		Board:   *b,
		BoardID: boardID,
	}
	got, err := newTemplateBoard(*b, boardID, okMockFreeSpacer)
	switch {
	case err != nil:
		t.Errorf("creating template board: %v", err)
	case len(got.FreeSpace) == 0:
		t.Errorf("board free space not set: %#v", got)
	default:
		got.FreeSpace = ""
		if !reflect.DeepEqual(want, got) {
			t.Errorf("template games not equal:\nwanted: %#v\ngot:    %#v", want, got)
		}
	}
}

func TestHandleHelp(t *testing.T) {
	var w bytes.Buffer
	err := executeHelpTemplate(&w)
	switch {
	case err != nil:
		t.Error(err)
	case w.Len() == 0:
		t.Error("no bytes written")
	}
}

func TestHandleAbout(t *testing.T) {
	var w bytes.Buffer
	err := executeAboutTemplate(&w)
	switch {
	case err != nil:
		t.Error(err)
	case w.Len() == 0:
		t.Error("no bytes written")
	}
}

func TestHandleGame(t *testing.T) {
	tests :=
		[]struct {
			name     string
			game     bingo.Game
			boardID  string
			hasBingo bool
			want     string
			negate   bool
		}{
			{
				name: "game has drawn tile",
				game: func() bingo.Game {
					var allDrawnGame bingo.Game
					prevNumbersLeft := allDrawnGame.NumbersLeft()
					for {
						allDrawnGame.DrawNumber()
						numbersLeft := allDrawnGame.NumbersLeft()
						if prevNumbersLeft == numbersLeft {
							return allDrawnGame
						}
						prevNumbersLeft = numbersLeft
					}
				}(),
				want: "B 2",
			},
			{
				name:    "checked board value",
				boardID: "board_id_input_value",
				want:    "board_id_input_value",
			},
			{
				name:     "checked board has no bingo",
				boardID:  "board_id_input_value",
				hasBingo: false,
				want:     "No Bingo :(</a>",
			},
			{
				name:     "checked board has bingo",
				boardID:  "board_id_input_value",
				hasBingo: true,
				want:     "BINGO !!!</a>",
			},
			{
				name: "has previous number",
				game: func() bingo.Game {
					var g bingo.Game
					g.DrawNumber()
					return g
				}(),
				want: "Previous number:",
			},
			{
				name:   "does not have previous number",
				game:   bingo.Game{},
				want:   "Previous number:",
				negate: true,
			},
		}
	for i, test := range tests {
		var w bytes.Buffer
		err := executeGameTemplate(&w, test.game, "game-id", test.boardID, test.hasBingo)
		got := w.String()
		switch {
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case strings.Contains(got, test.want) == test.negate:
			t.Errorf("test %v (%v): (negate-contains-check=%v): wanted rendered game to contain %q, got:\n%v", i, test.name, test.negate, test.want, got)
		}
	}
}

func TestHandleGames(t *testing.T) {
	var w bytes.Buffer
	gi := gameInfo{
		ID:          "1847",
		ModTime:     "time_text",
		NumbersLeft: 36,
	}
	gameInfos := []gameInfo{gi}
	err := executeGamesTemplate(&w, gameInfos)
	got := w.String()
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(got, gi.ID):
		t.Errorf("game ID missing: %v", got)
	case !strings.Contains(got, gi.ModTime):
		t.Errorf("game Modification Time missing: %v", got)
	case !strings.Contains(got, "36"):
		t.Errorf("game Numbers Left missing: %v", got)
	}
}

func TestHandleBoard(t *testing.T) {
	tests := []struct {
		FreeSpacer
		name   string
		wantOk bool
	}{
		{
			name:       "ok",
			wantOk:     true,
			FreeSpacer: okMockFreeSpacer,
		},
		{
			name:       "freeSpacer error",
			wantOk:     false,
			FreeSpacer: errMockFreeSpacer,
		},
		{
			name:   "png encode error - image with zero size",
			wantOk: false,
			FreeSpacer: &mockFreeSpacer{
				Image: image.NewGray16(image.Rect(0, 0, 0, 0)),
			},
		},
	}
	for i, test := range tests {
		var w bytes.Buffer
		b := board1257894001
		boardID := board1257894001ID
		err := executeBoardTemplate(&w, b, boardID, test.FreeSpacer)
		got := w.String()
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case !strings.Contains(got, boardID):
			t.Errorf("test %v (%v): board ID missing: %v", i, test.name, got)
		}
	}
}

func TestHandleExportBoard(t *testing.T) {
	tests := []struct {
		FreeSpacer
		name   string
		wantOk bool
	}{
		{
			name:       "ok",
			FreeSpacer: okMockFreeSpacer,
			wantOk:     true,
		},
		{
			name:       "bad",
			FreeSpacer: errMockFreeSpacer,
			wantOk:     false,
		},
	}
	for i, test := range tests {
		var w bytes.Buffer
		b := board1257894001
		boardID := board1257894001ID
		err := executeBoardExportTemplate(&w, b, boardID, test.FreeSpacer)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		default:
			got := w.String()
			for i, n := range b {
				if want := ">" + strconv.Itoa(int(n)) + "<"; i != 12 && !strings.Contains(got, want) {
					t.Errorf("wanted board export to contain %q:\n%v", want, got)
					break
				}
			}
		}
	}
}

func TestHandleIndexResponseWriter(t *testing.T) {
	tests := []struct {
		name string
		page
		t              *template.Template
		wantStatusCode int
		wantOk         bool
	}{
		{
			name:           "unknown template",
			t:              template.Must(template.New("unknown template").Parse("<p>template for {{.UNKNOWN}}</p>")),
			wantStatusCode: 500,
		},
		{
			name:           "empty game",
			page:           page{Name: "about"},
			t:              embeddedTemplate,
			wantStatusCode: 200,
			wantOk:         true,
		},
	}
	for i, test := range tests {
		w := httptest.NewRecorder()
		err := test.page.executeIndexTemplate(test.t, w)
		gotOk := err == nil
		gotStatusCode := w.Code
		switch {
		case test.wantOk != gotOk:
			t.Errorf("test %v (%v): ok values not equal: wanted %v, got %v (error: %v) ", i, test.name, test.wantOk, gotOk, err)
		case test.wantStatusCode != gotStatusCode:
			t.Errorf("test %v (%v): status codes not equal: wanted %v, got %v", i, test.name, test.wantStatusCode, gotStatusCode)
		}
	}
}

var board1257894001 = bingo.Board{
	15, 8, 4, 12, 10, // B
	19, 27, 16, 28, 25, // I
	42, 41, 0, 31, 40, // N
	49, 52, 50, 46, 57, // G
	64, 72, 67, 70, 74, // O
}
