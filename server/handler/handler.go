package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

type (
	// handler tracks servers HTTP requests and stores recent game infos.
	// The time function is used to create game infos
	handler struct {
		Mux
		gameInfos []gameInfo
		time      func() string
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
func Handler(gameCount int, time func() string) (http.Handler, error) {
	if gameCount < 1 {
		return nil, fmt.Errorf("positive GameCount required, got %v", gameCount)
	}
	if time == nil {
		return nil, fmt.Errorf("time function required")
	}
	h := handler{
		gameInfos: make([]gameInfo, 0, gameCount),
		time:      time,
	}
	return &h, nil
}

// Redirect is a handler that redirects all requests to HTTPS uris.
// The httpsPort is used to redirect requests to non-standard HTTPS ports.
func Redirect(httpsPort string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpsURI := "https://" + r.URL.Hostname()
		if len(r.URL.Port()) != 0 && httpsPort != "443" {
			httpsURI += ":" + httpsPort
		}
		httpsURI += r.URL.Path
		http.Redirect(w, r, httpsURI, http.StatusMovedPermanently)
	}
}

// ServeHTTP serves requests for GET and POST methods, not allowing others.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Mux == nil {
		h.Mux = Mux{
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
	h.Mux.ServeHTTP(w, r)
}

// httpError writes the message with statusCode to the response.
func httpError(w http.ResponseWriter, message string, statusCode int) {
	if len(message) == 0 {
		message = http.StatusText(statusCode)
	}
	http.Error(w, message, statusCode)
}

// getGames renders the games page onto the response with the game infos.
func (h handler) getGames(w http.ResponseWriter, r *http.Request) {
	executeGamesTemplate(w, h.gameInfos)
}

// getGame renders the game page onto the response with the game of the 'gameID' query parameter.
// The 'boardID' and 'bingo' query parameters are also used to forward the results of a BINGO check.
func (handler) getGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")
	boardID := r.URL.Query().Get("boardID")
	hasBingo := r.URL.Query().Has("bingo")
	var g *bingo.Game
	switch len(gameID) {
	case 0:
		g = new(bingo.Game)
	default:
		var ok bool
		g, ok = parseGame(gameID, w)
		if !ok {
			return
		}
	}
	executeGameTemplate(w, *g, boardID, hasBingo)
}

// createGame renders an empty game
func (handler) createGame(w http.ResponseWriter, r *http.Request) {
	var g bingo.Game
	var boardID string
	var hasBingo bool
	executeGameTemplate(w, g, boardID, hasBingo)
}

// getBoard renders the board page onto the response or create a new board and redirects to it.
func (handler) getBoard(w http.ResponseWriter, r *http.Request) {
	boardID := r.URL.Query().Get("boardID")
	b, ok := parseBoard(boardID, w)
	if !ok {
		return
	}
	executeBoardTemplate(w, *b)
}

// createBoard redirects to a new board.
func (handler) createBoard(w http.ResponseWriter, r *http.Request) {
	b := bingo.NewBoard()
	boardID, err := b.ID()
	if err != nil {
		message := fmt.Sprintf("problem getting new board id: %v\nboard: %#v", err, b)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/game/board?boardID="+boardID, http.StatusSeeOther)
}

// getHelp renders the help page onto the response.
func (handler) getHelp(w http.ResponseWriter, r *http.Request) {
	executeHelpTemplate(w)
}

// getAbout renders the about page onto the response.
func (handler) getAbout(w http.ResponseWriter, r *http.Request) {
	executeAboutTemplate(w)
}

// checkBoard checks the board on the game with a checkType using the 'gameID', 'boardID', and 'type' query parameters.
// The results of the check are included as query parameters onto a redirect to the game page.'
func (handler) checkBoard(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")
	g, ok := parseGame(gameID, w)
	if !ok {
		return
	}
	boardID := r.URL.Query().Get("boardID")
	b, ok := parseBoard(boardID, w)
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
		httpError(w, message, http.StatusBadRequest)
		return
	}
	url := fmt.Sprintf("/game?gameID=%v&boardID=%v", gameID, boardID)
	if result {
		url += "&bingo"
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}

// drawNumber draws a new number for the game specified by the request's 'gameID' form parameter.
// The response is redirected to the updated game.  It's updated state is stored in the game infos slice.
func (h *handler) drawNumber(w http.ResponseWriter, r *http.Request) {
	gameID := r.FormValue("gameID")
	g, ok := parseGame(gameID, w)
	if !ok {
		return
	}
	beforeNumsLeft := g.NumbersLeft()
	g.DrawNumber()
	afterNumsLeft := g.NumbersLeft()
	if beforeNumsLeft == afterNumsLeft {
		http.Redirect(w, r, "/game?gameID="+gameID, http.StatusNotModified)
		return
	}
	afterID, err := g.ID()
	if err != nil {
		message := fmt.Sprintf("unexpected problem getting id after drawing number from game with a VALID id %q: %v", gameID, err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
	if len(h.gameInfos) < cap(h.gameInfos) {
		h.gameInfos = append(h.gameInfos, gameInfo{}) // increase length
	}
	copy(h.gameInfos[1:], h.gameInfos) // shift right, overwriting last
	modTime := h.time()
	gi := gameInfo{
		ID:          afterID,
		ModTime:     modTime,
		NumbersLeft: afterNumsLeft,
	}
	h.gameInfos[0] = gi // set first
	http.Redirect(w, r, "/game?gameID="+gi.ID, http.StatusSeeOther)
}

// createBoards creates 'n' boards as specified by the request's form parameter, attaching the boards in a zip file.
func (handler) createBoards(w http.ResponseWriter, r *http.Request) {
	nQueryParam := r.FormValue("n")
	n, err := strconv.Atoi(nQueryParam)
	if err != nil {
		message := fmt.Sprintf("%v: example: /game/boards?n=5 creates 5 unique boards", err)
		httpError(w, message, http.StatusBadRequest)
		return
	}
	if n < 1 || n > 1000 {
		httpError(w, "n must be be between 1 and 1000", http.StatusBadRequest)
		return
	}
	var buf bytes.Buffer
	z := zip.NewWriter(&buf)
	for i := 1; i <= n; i++ {
		fileName := fmt.Sprintf("bingo_%v.svg", i)
		f, err := z.Create(fileName)
		if err != nil {
			message := fmt.Sprintf("unexpected problem creating file #%v in zip: %v", i, fileName)
			httpError(w, message, http.StatusInternalServerError)
			return
		}
		board := bingo.NewBoard()
		if err := executeBoardExportTemplate(f, *board); err != nil {
			message := fmt.Sprintf("unexpected problem adding board #%v to zip file: %v", i, err)
			httpError(w, message, http.StatusInternalServerError)
			return
		}
	}
	if err := z.Close(); err != nil {
		message := fmt.Sprintf("unexpected problem closing zip file: %v", err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
	w.Header().Set("Content-Disposition", "attachment; filename=bingo-boards.zip")
}

// parseGame parses the game, writing parse errors to the response
func parseGame(id string, w http.ResponseWriter) (g *bingo.Game, ok bool) {
	g, err := bingo.GameFromID(id)
	if err != nil {
		message := fmt.Sprintf("getting game from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
		return nil, false
	}
	return g, true
}

// parseBoard parses the board, writing parse errors to the response
func parseBoard(id string, w http.ResponseWriter) (b *bingo.Board, ok bool) {
	b, err := bingo.BoardFromID(id)
	if err != nil {
		message := fmt.Sprintf("getting board from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
		return nil, false
	}
	return b, true
}
