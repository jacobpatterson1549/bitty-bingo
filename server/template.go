package server

import (
	"embed"
	"html/template"
	"io"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

//go:embed templates
var templatesFS embed.FS

var t = template.Must(template.New("games.html").
	ParseFS(templatesFS, "templates/*"))

func handleExportBoard(w io.Writer, b bingo.Board) error {
	return t.Lookup("board.svg").Execute(w, b)
}

func handleHelp(w io.Writer) error {
	return t.Lookup("help.html").Execute(w, nil)
}

func handleAbout(w io.Writer) error {
	return t.Lookup("about.html").Execute(w, nil)
}

func handleGame(w io.Writer, g bingo.Game) error {
	return t.Lookup("game.html").Execute(w, g)
}

func handleGames(w io.Writer, gameInfos []gameInfo) error {
	return t.Lookup("games.html").Execute(w, gameInfos)
}
