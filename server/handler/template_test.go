package handler

import (
	"bytes"
	"errors"
	"html/template"
	"image"
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
		err := executeGameTemplate(&w, test.game, test.boardID, test.hasBingo)
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

// init applies a mock QRCode function to make tests IN THIS PACKAGE faster by not calling external libaries
func init() {
	qrCode = func(text string, width, height int) (image.Image, error) {
		r := image.Rect(0, 0, 1, 1)
		img := image.NewGray(r)
		return img, nil
	}
}

func TestFreeSpace(t *testing.T) {
	// according https://www.w3.org/TR/PNG/#5PNG-file-signature,
	//  -> all png files start with [137 80 78 71 13 10 26 10] which encodes to "iVBORw0KGgo"
	prevQREncoder := qrCode
	defer func() {
		qrCode = prevQREncoder
	}()
	const wantDataPrefix = "iVBORw0KGgo"
	tests := []struct {
		name   string
		board  bingo.Board
		wantOk bool
		genQR  qrEncoder
	}{
		{
			name: "invalid board: zero value",
		},
		{
			board:  *bingo.NewBoard(),
			name:   "random board",
			wantOk: true,
			genQR: func(text string, width, height int) (image.Image, error) {
				return image.NewGray(image.Rect(0, 0, 1, 1)), nil
			},
		},
		{
			board:  *bingo.NewBoard(),
			name:   "qr encode error",
			wantOk: false,
			genQR: func(text string, width, height int) (image.Image, error) {
				return nil, errors.New("qr code error")
			},
		},
		{
			board:  *bingo.NewBoard(),
			name:   "qr encode error",
			wantOk: false,
			genQR: func(text string, width, height int) (image.Image, error) {
				return nil, errors.New("qr code error")
			},
		},
		{
			board:  *bingo.NewBoard(),
			name:   "page encode error (empty image",
			wantOk: false,
			genQR: func(text string, width, height int) (image.Image, error) {
				return image.NewGray(image.Rect(0, 0, 0, 0)), nil
			},
		},
	}
	for i, test := range tests {
		qrCode = test.genQR
		data, err := freeSpace(test.board)
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case !strings.HasPrefix(data, wantDataPrefix):
			id, _ := test.board.ID() // ignore error because it did not cause an error generating the free space
			t.Errorf("test %v (%v): prefix of free space not equal for board of ID=%q:\n"+
				"the base64 encoding of the png image of the qr code of the board id was unwanted:\nwanted: %v\ngot:    %v",
				i, test.name, id, wantDataPrefix, data)
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
