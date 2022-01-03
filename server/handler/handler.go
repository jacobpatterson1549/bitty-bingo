// Package handler evaluates site requests.
package handler

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strconv"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

type (
	// BarCoder generates image of a bar code of the board, possibly with an external library.
	BarCoder interface {
		// BarCode encodes the board id to a bar code image with a width and height.
		BarCode(boardID string, width, height int) (image.Image, error)
	}
	// handler tracks servers HTTP requests and stores recent game infos.
	// The time function is used to create game infos
	handler struct {
		http.Handler
		BarCoder
		gameInfos []gameInfo
		time      func() string
		favicon   string
	}
	// gameInfo is the display value of the sate of a game at a specific time.
	gameInfo struct {
		// ID is the identifier of the game.
		ID string
		// ModTime is used to display when the game was last modified.
		ModTime string
		// NumbersLeft is the amount of Numbers that can still be drawn in the game.
		NumbersLeft int
	}
)

// Handler creates a HTTP handler to serve the site.
// The gameCount and time function are validated used from the config in the handler
// Responses are returned gzip compression when allowed.
func Handler(gameCount int, time func() string, b BarCoder) (http.Handler, error) {
	switch {
	case gameCount < 1:
		return nil, fmt.Errorf("positive GameCount required, got %v", gameCount)
	case time == nil:
		return nil, fmt.Errorf("time function required")
	case b == nil:
		return nil, fmt.Errorf("BarCoder required")
	}
	var faviconW bytes.Buffer
	if err := executeFaviconTemplate(&faviconW); err != nil {
		return nil, fmt.Errorf("creating favicon: %v", err)
	}
	faviconB := faviconW.Bytes()
	favicon := base64.StdEncoding.EncodeToString([]byte(faviconB))
	h := handler{
		gameInfos: make([]gameInfo, 0, gameCount),
		time:      time,
		BarCoder:  b,
		favicon:   favicon,
	}
	return &h, nil
}

// ServeHTTP serves requests for GET and POST methods, not allowing others.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Handler == nil {
		h.Handler = newMux(h)
	}
	h.Handler.ServeHTTP(w, r)
}

// newMux creates a new multiplexer to handle endpoints
func newMux(h *handler) *Mux {
	return &Mux{
		"GET": {
			"/":                 h.getGames,
			"/game":             h.getGame,
			"/game/board/check": h.checkBoard,
			"/game/board":       h.getBoard,
			"/help":             h.getHelp,
			"/about":            h.getAbout,
		},
		"POST": {
			"/game":             h.createGame,
			"/game/draw_number": h.drawNumber,
			"/game/board":       h.createBoard,
			"/game/boards":      h.createBoards,
		},
	}
}

// getGames renders the games page onto the response with the game infos.
func (h *handler) getGames(w http.ResponseWriter, r *http.Request) {
	executeGamesTemplate(w, h.favicon, h.gameInfos)
}

// getGame renders the game page onto the response with the game of the 'gameID' query parameter.
// The 'boardID' and 'bingo' query parameters are also used to forward the results of a BINGO check.
func (h handler) getGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")
	boardID := r.URL.Query().Get("boardID")
	hasBingo := r.URL.Query().Has("bingo")
	g, ok := h.parseGame(gameID, w)
	if !ok {
		return
	}
	executeGameTemplate(w, h.favicon, *g, gameID, boardID, hasBingo)
}

// createGame renders an empty game
func (h handler) createGame(w http.ResponseWriter, r *http.Request) {
	var g bingo.Game
	gameID, err := g.ID()
	if err != nil {
		message := fmt.Sprintf("unexpected problem getting new game id: %v=\ngame: %#v", err, g)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
	h.redirect(w, r, "/game?gameID="+gameID)
}

// getBoard renders the board page onto the response or create a new board and redirects to it.
func (h handler) getBoard(w http.ResponseWriter, r *http.Request) {
	boardID := r.URL.Query().Get("boardID")
	b, ok := h.parseBoard(boardID, w)
	if !ok {
		return
	}
	barCode, err := h.boardBarCode(boardID)
	if err != nil {
		message := fmt.Sprintf("unexpected problem creating board bar code: %v", err)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
	executeBoardTemplate(w, h.favicon, *b, boardID, barCode)
}

// createBoard redirects to a new board.
func (h handler) createBoard(w http.ResponseWriter, r *http.Request) {
	b := bingo.NewBoard()
	boardID, err := b.ID()
	if err != nil {
		message := fmt.Sprintf("unexpected problem getting new board id: %v\nboard: %#v", err, b)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
	h.redirect(w, r, "/game/board?boardID="+boardID)
}

// getHelp renders the help page onto the response.
func (h handler) getHelp(w http.ResponseWriter, r *http.Request) {
	executeHelpTemplate(w, h.favicon)
}

// getAbout renders the about page onto the response.
func (h handler) getAbout(w http.ResponseWriter, r *http.Request) {
	executeAboutTemplate(w, h.favicon)
}

// checkBoard checks the board on the game with a checkType using the 'gameID', 'boardID', and 'type' query parameters.
// The results of the check are included as query parameters onto a redirect to the game page.'
func (h handler) checkBoard(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")
	g, ok := h.parseGame(gameID, w)
	if !ok {
		return
	}
	boardID := r.URL.Query().Get("boardID")
	b, ok := h.parseBoard(boardID, w)
	if !ok {
		return
	}
	checkType := r.URL.Query().Get("type")
	var result bool
	switch checkType {
	case "HasLine":
		result = b.HasLine(*g)
	case "IsFilled":
		result = b.IsFilled(*g)
	default:
		message := fmt.Sprintf("unknown checkType %q", checkType)
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	url := fmt.Sprintf("/game?gameID=%v&boardID=%v", gameID, boardID)
	if result {
		url += "&bingo"
	}
	h.redirect(w, r, url)
}

// drawNumber draws a new number for the game specified by the request's 'gameID' form parameter.
// The response is redirected to the updated game.  It's updated state is stored in the game infos slice.
func (h *handler) drawNumber(w http.ResponseWriter, r *http.Request) {
	gameID := r.FormValue("gameID")
	g, ok := h.parseGame(gameID, w)
	if !ok {
		return
	}
	beforeNumsLeft := g.NumbersLeft()
	g.DrawNumber()
	afterNumsLeft := g.NumbersLeft()
	if beforeNumsLeft == afterNumsLeft {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	afterID, err := g.ID()
	if err != nil {
		message := fmt.Sprintf("unexpected problem getting id after drawing number from game with a VALID id %q: %v", gameID, err)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
	h.addGame(afterID, afterNumsLeft)
	h.redirect(w, r, "/game?gameID="+afterID)
}

// addGame creates a new gameInfo and adds it to the gameInfos stack.  If the stack is full, the last item is discarded.
func (h *handler) addGame(gameID string, numbersLeft int) {
	if len(h.gameInfos) < cap(h.gameInfos) {
		h.gameInfos = append(h.gameInfos, gameInfo{}) // increase length
	}
	copy(h.gameInfos[1:], h.gameInfos) // shift right, overwriting last
	modTime := h.time()
	gi := gameInfo{
		ID:          gameID,
		ModTime:     modTime,
		NumbersLeft: numbersLeft,
	}
	h.gameInfos[0] = gi // set first
}

// createBoards creates 'n' boards as specified by the request's form parameter, attaching the boards in a zip file.
func (h handler) createBoards(w http.ResponseWriter, r *http.Request) {
	nQueryParam := r.FormValue("n")
	n, err := strconv.Atoi(nQueryParam)
	if err != nil {
		message := fmt.Sprintf("%v: example: /game/boards?n=5 creates 5 unique boards", err)
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	if n < 1 || n > 1000 {
		http.Error(w, "n must be be between 1 and 1000", http.StatusBadRequest)
		return
	}
	var buf bytes.Buffer
	if err := h.zipNewBoards(&buf, n); err != nil {
		message := fmt.Sprintf("unexpected problem creating zip file: %v", err)
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
	w.Header().Set("Content-Disposition", "attachment; filename=bingo-boards.zip")
}

// zipNewBoards writes n new boards to a zip file
func (h handler) zipNewBoards(w io.Writer, n int) error {
	z := zip.NewWriter(w)
	for i := 1; i <= n; i++ {
		fileName := fmt.Sprintf("bingo_%v.svg", i)
		f, err := z.Create(fileName)
		if err != nil {
			return fmt.Errorf("creating file #%v: %v", i, fileName)
		}
		b := bingo.NewBoard()
		boardID, err := b.ID()
		if err != nil {
			return fmt.Errorf("getting id of board #%v: %v\nboard: %#v", i, err, b)
		}
		barCode, err := h.boardBarCode(boardID)
		if err != nil {
			return fmt.Errorf("creating board #%v bar code: %v", i, err)
		}
		if err := executeBoardExportTemplate(f, *b, boardID, barCode); err != nil {
			return fmt.Errorf("adding board #%v to zip file: %v", i, err)
		}
	}
	if err := z.Close(); err != nil {
		return fmt.Errorf("writing/closing zip file: %v", err)
	}
	return nil
}

// parseGame parses the game, writing parse errors to the response
func (h handler) parseGame(id string, w http.ResponseWriter) (g *bingo.Game, ok bool) {
	g, err := bingo.GameFromID(id)
	if err != nil {
		message := fmt.Sprintf("getting game from query parameter: %v", err)
		http.Error(w, message, http.StatusBadRequest)
		return nil, false
	}
	return g, true
}

// parseBoard parses the board, writing parse errors to the response
func (h handler) parseBoard(id string, w http.ResponseWriter) (b *bingo.Board, ok bool) {
	b, err := bingo.BoardFromID(id)
	if err != nil {
		message := fmt.Sprintf("getting board from query parameter: %v", err)
		http.Error(w, message, http.StatusBadRequest)
		return nil, false
	}
	return b, true
}

// boardBarCode uses the BarCoder to encode the bar code image as a base64-encode png image with transparency.
func (h handler) boardBarCode(boardID string) (string, error) {
	barCode, err := h.BarCoder.BarCode(boardID, 80, 80)
	if err != nil {
		return "", fmt.Errorf("creating bar code: %v", err)
	}
	var buf bytes.Buffer
	img := newTransparentImage(barCode)
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("bar code to png image: %v", err)
	}
	bytes := buf.Bytes()
	data := base64.StdEncoding.EncodeToString(bytes)
	return data, nil
}

// redirect tells the response to see a different url.
func (h handler) redirect(w http.ResponseWriter, r *http.Request, url string) {
	http.Redirect(w, r, url, http.StatusSeeOther)
}
