package server

import (
	"embed"
	"html/template"
	"io"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

//go:embed templates
var templatesFS embed.FS

var t = template.Must(template.ParseFS(templatesFS, "templates/*"))

type page struct {
	Name string
	List []gameInfo
	Game *bingo.Game
}

func handleHelp(w io.Writer) error {
	return t.ExecuteTemplate(w, "index.html", page{Name: "help"})
}

func handleAbout(w io.Writer) error {
	return t.ExecuteTemplate(w, "index.html", page{Name: "about"})
}

func handleGame(w io.Writer, g *bingo.Game) error {
	return t.ExecuteTemplate(w, "index.html", page{Name: "game", Game: g})
}

func handleGames(w io.Writer, gameInfos []gameInfo) error {
	return t.ExecuteTemplate(w, "index.html", page{Name: "list", List: gameInfos})
}

func handleExportBoard(w io.Writer, b bingo.Board) error {
	return t.ExecuteTemplate(w, "board.svg", b)
}
