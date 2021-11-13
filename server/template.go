package server

import (
	"embed"
	"html/template"
	"io"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

// templateFS is the embedded filesystem containing the template files.
//go:embed templates
var templatesFS embed.FS

// t is the template containing the html and svg templates.
var t = template.Must(template.ParseFS(templatesFS, "templates/*"))

// page contains all the data needed to render any html page.
type page struct {
	Name string
	List []gameInfo
	Game *game
}

// game contains the fields to render a game page.
type game struct {
	bingo.Game
	BoardID  string
	HasBingo bool
}

// handleHelp renders the help html page.
func handleHelp(w io.Writer) error {
	return t.ExecuteTemplate(w, "index.html", page{Name: "help"})
}

// handleAbout renders the about html page.
func handleAbout(w io.Writer) error {
	return t.ExecuteTemplate(w, "index.html", page{Name: "about"})
}

// handleGame renders the game html page.
func handleGame(w io.Writer, g *bingo.Game, boardID string, hasBingo bool) error {
	templateGame := game{
		Game:     *g,
		BoardID:  boardID,
		HasBingo: hasBingo,
	}
	return t.ExecuteTemplate(w, "index.html", page{Name: "game", Game: &templateGame})
}

// handleGames renders the games list html page.
func handleGames(w io.Writer, gameInfos []gameInfo) error {
	return t.ExecuteTemplate(w, "index.html", page{Name: "list", List: gameInfos})
}

// handleExportBoard renders the board onto an svg image.
func handleExportBoard(w io.Writer, b bingo.Board) error {
	return t.ExecuteTemplate(w, "board.svg", b)
}
