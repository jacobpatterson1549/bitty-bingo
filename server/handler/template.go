package handler

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"image"
	"io"
	"net/http"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
	"github.com/jacobpatterson1549/bitty-bingo/server/handler/qr"
)

var (
	// templatesFS is the embedded filesystem containing the template files.
	//go:embed templates
	templatesFS embed.FS
	// embeddedTemplate is the template containing the html and svg templates.
	embeddedTemplate = template.Must(template.ParseFS(templatesFS, "templates/*"))
	// Image creates an QR codeimage from the text.  For testing use only.
	qrCode qrEncoder = qr.Image
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
		Board     bingo.Board
		BoardID   string
		FreeSpace string
	}
	// qrEncoder encodes text to a QR code with the specified size.
	qrEncoder func(text string, width, height int) (image.Image, error)
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
func newTemplateBoard(b bingo.Board, id string) (*board, error) {
	data, err := freeSpace(b)
	if err != nil {
		return nil, fmt.Errorf("getting center square free space for board: %v", err)
	}
	templateBoard := board{
		Board:     b,
		BoardID:   id,
		FreeSpace: data,
	}
	return &templateBoard, nil
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
func executeBoardTemplate(w io.Writer, b bingo.Board, boardID string) error {
	templateBoard, err := newTemplateBoard(b, boardID)
	if err != nil {
		return err
	}
	p := page{
		Name:  "board",
		Board: templateBoard,
	}
	return p.executeIndexTemplate(embeddedTemplate, w)
}

// executeBoardExportTemplate renders the board onto an svg image.
func executeBoardExportTemplate(w io.Writer, b bingo.Board, boardID string) error {
	templateBoard, err := newTemplateBoard(b, boardID)
	if err != nil {
		return err
	}
	return embeddedTemplate.ExecuteTemplate(w, "board.svg", templateBoard)
}

// executeIndexTemplate renders the page on the index HTML template.
// HTTPErrors are handled if Writer is a ResponseWriter.
// Templates are written a buffer to ensure they execute correctly before they are written to the response
func (p page) executeIndexTemplate(t *template.Template, w io.Writer) error {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "index.html", p); err != nil {
		if rw, ok := w.(http.ResponseWriter); ok {
			message := fmt.Sprintf("unexpected problem rendering %v template: %v", p.Name, err)
			httpError(rw, message, http.StatusInternalServerError)
		}
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}
