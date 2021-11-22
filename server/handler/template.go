package handler

import (
	"embed"
	"html/template"
	"io"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

var (
	// templatesFS is the embedded filesystem containing the template files.
	//go:embed templates
	templatesFS embed.FS
	// embeddedTemplate is the template containing the html and svg templates.
	embeddedTemplate = template.Must(template.ParseFS(templatesFS, "templates/*"))
)

type (
	// page contains all the data needed to render any html page.
	page struct {
		Name  string
		List  []gameInfo
		Game  *game
		Board *board
	}
	// game contains the fields to render a game page.
	game struct {
		Game     bingo.Game
		GameID   string
		BoardID  string
		HasBingo bool
	}
	// board contains the field to export a board
	board struct {
		Board   bingo.Board
		BoardID string
		// FreeSpace is the base64 encoded png image that should be placed in the free space in the middle of the board
		FreeSpace string
	}
)

func newTemplateGame(g bingo.Game, gameID, boardID string, hasBingo bool) *game {
	templateGame := game{
		Game:     g,
		GameID:   gameID,
		BoardID:  boardID,
		HasBingo: hasBingo,
	}
	return &templateGame
}

// newTemplateBoard creates a board to render from the bingo board
func newTemplateBoard(b bingo.Board, boardID, freeSpace string) *board {
	templateBoard := board{
		Board:     b,
		BoardID:   boardID,
		FreeSpace: freeSpace,
	}
	return &templateBoard
}

// executeHelpTemplate renders the help html page.
func executeHelpTemplate(w io.Writer) error {
	p := page{
		Name: "help",
	}
	return p.executeIndexTemplate(embeddedTemplate, w)
}

// executeAboutTemplate renders the about html page.
func executeAboutTemplate(w io.Writer) error {
	p := page{
		Name: "about",
	}
	return p.executeIndexTemplate(embeddedTemplate, w)
}

// executeGameTemplate renders the game html page.
func executeGameTemplate(w io.Writer, g bingo.Game, gameID, boardID string, hasBingo bool) error {
	templateGame := newTemplateGame(g, gameID, boardID, hasBingo)
	p := page{
		Name: "game",
		Game: templateGame,
	}
	return p.executeIndexTemplate(embeddedTemplate, w)
}

// executeGamesTemplate renders the games list html page.
func executeGamesTemplate(w io.Writer, gameInfos []gameInfo) error {
	p := page{
		Name: "list",
		List: gameInfos,
	}
	return p.executeIndexTemplate(embeddedTemplate, w)
}

// executeBoardTemplate renders the board on the html page.
func executeBoardTemplate(w io.Writer, b bingo.Board, boardID, freeSpace string) error {
	templateBoard := newTemplateBoard(b, boardID, freeSpace)
	p := page{
		Name:  "board",
		Board: templateBoard,
	}
	return p.executeIndexTemplate(embeddedTemplate, w)
}

// executeBoardExportTemplate renders the board onto an svg image.
func executeBoardExportTemplate(w io.Writer, b bingo.Board, boardID, freeSpace string) error {
	templateBoard := newTemplateBoard(b, boardID, freeSpace)
	return embeddedTemplate.ExecuteTemplate(w, "board.svg", templateBoard)
}

// executeIndexTemplate renders the page on the index HTML template.
func (p page) executeIndexTemplate(t *template.Template, w io.Writer) error {
	return t.ExecuteTemplate(w, "index.html", p)
}
