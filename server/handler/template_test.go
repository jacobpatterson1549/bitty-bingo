package handler

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

func TestHandleHelp(t *testing.T) {
	var w bytes.Buffer
	err := executeHelpTemplate(&w, "FAVICON-1")
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(w.String(), "FAVICON-1"):
		t.Error("wanted page to contain favicon")
	}
}

func TestHandleAbout(t *testing.T) {
	var w bytes.Buffer
	err := executeAboutTemplate(&w, "FAVICON-2")
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(w.String(), "FAVICON-2"):
		t.Errorf("wanted page to contain FAVICON-2: %v", w.String())
	}
}

func TestHandleGame(t *testing.T) {
	var allNumbersDrawnGame bingo.Game
	for allNumbersDrawnGame.NumbersLeft() != 0 {
		allNumbersDrawnGame.DrawNumber()
	}
	var oneNumberDrawnGame bingo.Game
	oneNumberDrawnGame.DrawNumber()
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
				game: allNumbersDrawnGame,
				want: "B 2",
			},
			{
				name:    "checked board value",
				game:    oneNumberDrawnGame,
				boardID: "board_id_input_value",
				want:    "board_id_input_value",
			},
			{
				name:     "checked board has no bingo",
				game:     oneNumberDrawnGame,
				boardID:  "board_id_input_value",
				hasBingo: false,
				want:     "No Bingo :(</a>",
			},
			{
				name:     "checked board has bingo",
				game:     oneNumberDrawnGame,
				boardID:  "board_id_input_value",
				hasBingo: true,
				want:     "BINGO !!!</a>",
			},
			{
				name: "has previous number",
				game: oneNumberDrawnGame,
				want: "Previous number:",
			},
			{
				name:   "does not have previous number",
				game:   bingo.Game{},
				want:   "Previous number:",
				negate: true,
			},
			{
				name:   "does not have check board when game is empty",
				game:   bingo.Game{},
				want:   "Check Board",
				negate: true,
			},
			{
				name: "has check board when game is started",
				game: oneNumberDrawnGame,
				want: "Check Board",
			},
		}
	for i, test := range tests {
		var w bytes.Buffer
		err := executeGameTemplate(&w, "FAVICON-3", test.game, "game-id", test.boardID, test.hasBingo)
		got := w.String()
		switch {
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case !strings.Contains(w.String(), "FAVICON-3"):
			t.Errorf("wanted page to contain FAVICON-3: %v", w.String())
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
	err := executeGamesTemplate(&w, "FAVICON-4", gameInfos)
	got := w.String()
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(w.String(), "FAVICON-4"):
		t.Errorf("wanted page to contain FAVICON-4: %v", w.String())
	case !strings.Contains(got, gi.ID):
		t.Errorf("game ID missing: %v", got)
	case !strings.Contains(got, gi.ModTime):
		t.Errorf("game Modification Time missing: %v", got)
	case !strings.Contains(got, "36"):
		t.Errorf("game Numbers Left missing: %v", got)
	}
}

func TestHandleBoard(t *testing.T) {
	var w bytes.Buffer
	var b bingo.Board
	boardID := "board-313"
	freeSpace := "free-space-png-base64-data"
	err := executeBoardTemplate(&w, "FAVICON-5", b, boardID, freeSpace)
	got := w.String()
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(w.String(), "FAVICON-5"):
		t.Errorf("wanted page to contain FAVICON-5: %v", w.String())
	case !strings.Contains(got, boardID):
		t.Errorf("board ID missing: %v", got)
	}
}

func TestHandleBoardExport(t *testing.T) {
	var w bytes.Buffer
	var b bingo.Board
	boardID := "board-313"
	freeSpace := "free-space-png-base64-data-2"
	err := executeBoardExportTemplate(&w, b, boardID, freeSpace)
	got := w.String()
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(got, boardID):
		t.Errorf("board ID missing: %v", got)
	case !strings.Contains(got, freeSpace):
		t.Errorf("board free space missing: %v", got)
	}
}

func TestHandleFavicon(t *testing.T) {
	var w bytes.Buffer
	err := executeFaviconTemplate(&w)
	got := w.String()
	wantPrefix := "<svg"
	switch {
	case err != nil:
		t.Error(err)
	case !strings.HasPrefix(got, wantPrefix):
		t.Errorf("wanted favicon to start with %q:\n%v", wantPrefix, got)
	}
}
