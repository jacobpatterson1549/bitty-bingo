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
	Funcs(template.FuncMap{
		"Value": func(n bingo.Number) int {
			return n.Value()
		},
	}).
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
	cols := make(map[int][]bingo.Number, 5)
	for _, n := range g.DrawnNumbers {
		cols[n.Column()] = append(cols[n.Column()], n)
	}
	return t.Lookup("game.html").Execute(w, cols)
}

func handleGames(w io.Writer, games []bingo.Game) error {
	return t.Lookup("games.html").Execute(w, games)
}
