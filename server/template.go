package server

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

//go:embed templates
var templatesFS embed.FS

var t = template.Must(template.New("games.html").
	ParseFS(templatesFS, "templates/*"))

func handleHelp(w io.Writer) error {
	return executeTemplate("help.html", w, nil)
}

func handleAbout(w io.Writer) error {
	return executeTemplate("about.html", w, nil)
}

func handleGame(w io.Writer, g bingo.Game) error {
	return executeTemplate("game.html", w, g)
}

func handleGames(w io.Writer, gameInfos []gameInfo) error {
	return executeTemplate("games.html", w, gameInfos)
}

func handleExportBoard(w io.Writer, b bingo.Board) error {
	return executeTemplate("board.svg", w, b)
}

func executeTemplate(name string, w io.Writer, data interface{}) error {
	t := t.Lookup(name)
	if t == nil {
		return fmt.Errorf("no template named %q", name)
	}
	return t.Execute(w, data)
}
