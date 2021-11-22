package handler

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
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
	freeSpace := "free-space-png-base64-data"
	want := &board{
		Board:     *b,
		BoardID:   boardID,
		FreeSpace: freeSpace,
	}
	got := newTemplateBoard(*b, boardID, freeSpace)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("template games not equal:\nwanted: %#v\ngot:    %#v", want, got)
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
	var w bytes.Buffer
	b := board1257894001
	boardID := board1257894001ID
	freeSpace := "free-space-png-base64-data"
	err := executeBoardTemplate(&w, b, boardID, freeSpace)
	got := w.String()
	switch {
	case err != nil:
		t.Errorf("unwanted error: %v", err)
	case !strings.Contains(got, boardID):
		t.Errorf("board ID missing: %v", got)
	}
}

var board1257894001 = bingo.Board{
	15, 8, 4, 12, 10, // B
	19, 27, 16, 28, 25, // I
	42, 41, 0, 31, 40, // N
	49, 52, 50, 46, 57, // G
	64, 72, 67, 70, 74, // O
}
