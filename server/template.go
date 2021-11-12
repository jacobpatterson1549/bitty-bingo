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

func handleHelp(w io.Writer) error {
	return t.ExecuteTemplate(w, "help.html", nil)
}

func handleAbout(w io.Writer) error {
	return t.ExecuteTemplate(w, "about.html", nil)
}

func handleGame(w io.Writer, g bingo.Game) error {
	return t.ExecuteTemplate(w, "game.html", g)
}

func handleGames(w io.Writer, gameInfos []gameInfo) error {
	return t.ExecuteTemplate(w, "games.html", gameInfos)
}

func handleExportBoard(w io.Writer, b bingo.Board) error {
	return t.ExecuteTemplate(w, "board.svg", b)
}
