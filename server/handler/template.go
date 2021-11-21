package handler

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"image"
	"image/png"
	"io"
	"net/http"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

var (
	// templatesFS is the embedded filesystem containing the template files.
	//go:embed templates
	templatesFS embed.FS
	// embeddedTemplate is the template containing the html and svg templates.
	embeddedTemplate = template.Must(template.ParseFS(templatesFS, "templates/*"))
	// Image creates an QR codeimage from the text.  For testing use only.
)

type (
	// FreeSpacer generates the free space image for a board.
	FreeSpacer interface {
		// QRCode encodes the board id to a QR code with a width and height.
		QRCode(boardID string, width, height int) (image.Image, error)
	}
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
func newTemplateBoard(b bingo.Board, boardID string, freeSpacer FreeSpacer) (*board, error) {
	qrCode, err := freeSpacer.QRCode(boardID, 80, 80)
	if err != nil {
		return nil, fmt.Errorf("creating qr code for free space: %v", err)
	}
	var buf bytes.Buffer
	img := newTransparentImage(qrCode)
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("converting free space qr code to png image: %v", err)
	}
	bytes := buf.Bytes()
	data := base64.StdEncoding.EncodeToString(bytes)
	templateBoard := board{
		Board:     b,
		BoardID:   boardID,
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
func executeBoardTemplate(w io.Writer, b bingo.Board, boardID string, freeSpacer FreeSpacer) error {
	templateBoard, err := newTemplateBoard(b, boardID, freeSpacer)
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
func executeBoardExportTemplate(w io.Writer, b bingo.Board, boardID string, freeSpacer FreeSpacer) error {
	templateBoard, err := newTemplateBoard(b, boardID, freeSpacer)
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
