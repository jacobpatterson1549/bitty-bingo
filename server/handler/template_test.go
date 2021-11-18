package handler

import (
	"bytes"
	"html/template"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

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
		}{
			{
				name: "game has drawn tile",
				game: *allNumbersDrawnGame(t),
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
		}
	for i, test := range tests {
		var w bytes.Buffer
		err := executeGameTemplate(&w, test.game, test.boardID, test.hasBingo)
		got := w.String()
		switch {
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case !strings.Contains(got, test.want):
			t.Errorf("test %v (%v): wanted rendered game to contain %q, got:\n%v", i, test.name, test.want, got)
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
	t.Run("ok", func(t *testing.T) {
		var w bytes.Buffer
		b := board1257894001
		err := executeBoardTemplate(&w, b)
		got := w.String()
		switch {
		case err != nil:
			t.Error(err)
		case !strings.Contains(got, board1257894001ID):
			t.Errorf("board ID missing: %v", got)
		}
	})
	t.Run("bad", func(t *testing.T) {
		var w bytes.Buffer
		var b bingo.Board
		if err := executeBoardTemplate(&w, b); err == nil {
			t.Error("wanted export error rending board with bad id")
		}
	})
}

func TestHandleExportBoard(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		b := board1257894001
		var w bytes.Buffer
		if err := executeBoardExportTemplate(&w, b); err != nil {
			t.Fatalf("unwanted export error: %v", err)
		}
		got := w.String()
		for i, n := range b {
			if want := ">" + strconv.Itoa(int(n)) + "<"; i != 12 && !strings.Contains(got, want) {
				t.Errorf("wanted board export to contain %q:\n%v", want, got)
				break
			}
		}
	})
	t.Run("bad", func(t *testing.T) {
		var b bingo.Board
		var w bytes.Buffer
		if err := executeBoardExportTemplate(&w, b); err == nil {
			t.Error("wanted export error exporting board with bad id")
		}
	})
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

func allNumbersDrawnGame(t *testing.T) *bingo.Game {
	t.Helper()
	var allDrawnGame bingo.Game
	prevNumbersLeft := allDrawnGame.NumbersLeft()
	for {
		allDrawnGame.DrawNumber()
		numbersLeft := allDrawnGame.NumbersLeft()
		if prevNumbersLeft == numbersLeft {
			return &allDrawnGame
		}
		prevNumbersLeft = numbersLeft
	}
}