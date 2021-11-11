package server

import (
	"bytes"
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
	var w bytes.Buffer
	var g bingo.Game
	for i := 0; i < int(bingo.MaxNumber); i++ {
		g.DrawNumber()
	}
	err := handleGame(&w, g)
	got := w.String()
	switch {
	case err != nil:
		t.Error(err)
	case !strings.Contains(got, "B 2"):
		t.Errorf("wanted B 2:\n%v", got)
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

func TestExecuteTemplate_badName(t *testing.T) {
	var w bytes.Buffer
	gotErr := executeTemplate("UNKNOWN", &w, nil)
	if gotErr == nil {
		t.Errorf("wanted error executing template with unknown name")
	}
}
