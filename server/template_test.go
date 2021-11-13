package server

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
	err := handleHelp(&w)
	switch {
	case err != nil:
		t.Error(err)
	case w.Len() == 0:
		t.Error("no bytes written")
	}
}

func TestHandleAbout(t *testing.T) {
	var w bytes.Buffer
	err := handleAbout(&w)
	switch {
	case err != nil:
		t.Error(err)
	case w.Len() == 0:
		t.Error("no bytes written")
	}
}

func TestHandleGame(t *testing.T) {
	var allDrawnGame bingo.Game
	for i := 0; i < int(bingo.MaxNumber); i++ {
		allDrawnGame.DrawNumber()
	}
	tests :=
		[]struct {
			game     bingo.Game
			boardID  string
			hasBingo bool
			want     string
		}{
			{},
			{
				game: allDrawnGame,
				want: "B 2",
			},
			{
				boardID: "board_id_input_value",
				want:    "board_id_input_value",
			},
			{
				boardID:  "board_id_input_value",
				hasBingo: false,
				want:     "No Bingo :(",
			},
			{
				boardID:  "board_id_input_value",
				hasBingo: true,
				want:     "BINGO !!!",
			},
		}
	for i, test := range tests {
		var w bytes.Buffer
		err := handleGame(&w, test.game, test.boardID, test.hasBingo)
		got := w.String()
		switch {
		case err != nil:
			t.Errorf("test %v: unwanted error: %v", i, err)
		case !strings.Contains(got, test.want):
			t.Errorf("test %v: wanted rendered game to contain %q, got:\n%v", i, test.want, got)
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
	err := handleGames(&w, gameInfos)
	got := w.String()
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(got, gi.ID):
		t.Errorf("game ID missing:\n%v", got)
	case !strings.Contains(got, gi.ModTime):
		t.Errorf("game Modification Time missing:\nn%v", got)
	case !strings.Contains(got, "36"):
		t.Errorf("game Numbers Left missing:\n%v", got)
	}
}

func TestHandleExport(t *testing.T) {
	b := bingo.Board{
		15, 8, 4, 12, 10, // B
		19, 27, 16, 28, 25, // I
		42, 41, 0, 31, 40, // N
		49, 52, 50, 46, 57, // G
		64, 72, 67, 70, 74, // O
	}
	var w bytes.Buffer
	gotErr := handleExportBoard(&w, b)
	if gotErr != nil {
		t.Fatalf("unwanted export error: %v", gotErr)
	}
	got := w.String()
	for i, n := range b {
		if i == 12 {
			continue
		}
		s := ">" + strconv.Itoa(int(n)) + "<"
		if !strings.Contains(got, s) {
			t.Errorf("wanted board export to contain %q:\n%v", s, got)
			break
		}
	}
}

func TestHandleIndexResponseWriter(t *testing.T) {
	tests := []struct {
		page
		t              *template.Template
		wantStatusCode int
		wantOk         bool
	}{
		{ // unknown template
			t:              template.Must(template.New("unknown template").Parse("<p>template for {{.UNKNOWN}}</p>")),
			wantStatusCode: 500,
		},
		{ // empty game
			page:           page{Name: "about"},
			t:              embeddedTemplate,
			wantStatusCode: 200,
			wantOk:         true,
		},
	}
	for i, test := range tests {
		w := httptest.NewRecorder()
		err := test.page.handleIndex(test.t, w)
		gotOk := err == nil
		gotStatusCode := w.Code
		switch {
		case test.wantOk != gotOk:
			t.Errorf("test %v: ok values not equal: wanted %v, got %v (error: %v) ", i, test.wantOk, gotOk, err)
		case test.wantStatusCode != gotStatusCode:
			t.Errorf("test %v: status codes not equal: wanted %v, got %v", i, test.wantStatusCode, gotStatusCode)
		}
	}
}
