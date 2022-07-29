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

const (
	indexTemplateName       = "index.html"
	faviconTemplateName     = "favicon.svg"
	boardExportTemplateName = "board.svg"
)

type (
	// page contains all the data needed to render any html page.
	page struct {
		Name    string
		Favicon string
	}
	// gamesPage contains the fields to render a list of games
	gamesPage struct {
		page
		List []gameInfo
	}
	// gamePage contains the fields to render a gamePage page.
	gamePage struct {
		page
		Game     bingo.Game
		GameID   string
		BoardID  string
		HasBingo bool
	}
	// boardPage contains the field to export a boardPage
	boardPage struct {
		page
		Board   bingo.Board
		BoardID string
		// Barcode is a base64 encoded png image of a bar code that should be placed in the free space in the middle of the board
		Barcode string
	}
)

// executeHelpTemplate renders the help html page.
func executeHelpTemplate(w io.Writer, favicon string) error {
	p := page{
		Name:    "help",
		Favicon: favicon,
	}
	return embeddedTemplate.ExecuteTemplate(w, indexTemplateName, p)
}

// executeAboutTemplate renders the about html page.
func executeAboutTemplate(w io.Writer, favicon string) error {
	p := page{
		Name:    "about",
		Favicon: favicon,
	}
	return embeddedTemplate.ExecuteTemplate(w, indexTemplateName, p)
}

// executeGameTemplate renders the game html page.
func executeGameTemplate(w io.Writer, favicon string, g bingo.Game, gameID, boardID string, hasBingo bool) error {
	p := gamePage{
		page: page{
			Name:    "game",
			Favicon: favicon,
		},
		Game:     g,
		GameID:   gameID,
		BoardID:  boardID,
		HasBingo: hasBingo,
	}
	return embeddedTemplate.ExecuteTemplate(w, indexTemplateName, p)
}

// executeGamesTemplate renders the games list html page.
func executeGamesTemplate(w io.Writer, favicon string, gameInfos []gameInfo) error {
	p := gamesPage{
		page: page{
			Name:    "list",
			Favicon: favicon,
		},
		List: gameInfos,
	}
	return embeddedTemplate.ExecuteTemplate(w, indexTemplateName, p)
}

// executeBoardTemplate renders the board on the html page.
func executeBoardTemplate(w io.Writer, favicon string, b bingo.Board, boardID, barcode string) error {
	p := boardPage{
		page: page{
			Name:    "board",
			Favicon: favicon,
		},
		Board:   b,
		BoardID: boardID,
		Barcode: barcode,
	}
	return embeddedTemplate.ExecuteTemplate(w, indexTemplateName, p)
}

// executeBoardExportTemplate renders the board onto an svg image.
func executeBoardExportTemplate(w io.Writer, b bingo.Board, boardID, barcode string) error {
	data := boardPage{
		Board:   b,
		BoardID: boardID,
		Barcode: barcode,
	}
	return embeddedTemplate.ExecuteTemplate(w, boardExportTemplateName, data)
}

// executeFaviconTemplate renders the favicon without line breaks.
func executeFaviconTemplate(w io.Writer) error {
	return embeddedTemplate.ExecuteTemplate(w, faviconTemplateName, nil)
}
