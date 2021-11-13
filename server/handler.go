package server

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

// httpHandler creates a HTTP handler that redirects all requests to HTTPS.
// Responses are returned gzip compression when allowed.
func (cfg Config) httpHandler() http.Handler {
	h := httpsRedirectHandler(cfg.HTTPSPort)
	return withGzip(h)
}

// httpsHandler creates a HTTP handler to serve the site.
// The gameCount and time function are validated used from the config in the handler
// Responses are returned gzip compression when allowed.
func (cfg Config) httpsHandler() (http.Handler, error) {
	if cfg.GameCount < 1 {
		return nil, fmt.Errorf("positive GameCount required, got %v", cfg.GameCount)
	}
	if cfg.Time == nil {
		return nil, fmt.Errorf("time function required")
	}
	h := httpsHandler{
		gameInfos: make([]gameInfo, 0, cfg.GameCount),
		time:      cfg.Time,
	}
	return withGzip(&h), nil
}

// httpsRedirectHandler is a handler that redirects all requests to HTTPS uris.
// The httpsPort is used to redirect requests to non-standard HTTPS ports.
func httpsRedirectHandler(httpsPort string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpsURI := "https://" + r.URL.Hostname()
		if len(r.URL.Port()) != 0 && httpsPort != "443" {
			httpsURI += ":" + httpsPort
		}
		httpsURI += r.URL.Path
		http.Redirect(w, r, httpsURI, http.StatusMovedPermanently)
	}
}

// gameInfo is the display value of the sate of a game at a specific time.
type gameInfo struct {
	// ID is the identifier of the game.
	ID string
	// ModTime is used to display when the game was last modified.
	ModTime string
	// NumbersLeft is the amount of Numbers that can still be drawn in the game.
	NumbersLeft int
}

// httpsHandler tracks servers HTTP requests and stores recent game infos.
// The time function is used to create game infos
type httpsHandler struct {
	gameInfos []gameInfo
	time      func() string
}

// ServeHTTP serves requests for GET and POST methods, not allowing others.
func (h *httpsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.serveGet(w, r)
	case http.MethodPost:
		h.servePost(w, r)
	default:
		httpError(w, "", http.StatusMethodNotAllowed)
	}
}

// serveGet handles various GET requests.
func (h httpsHandler) serveGet(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.getGames(w, r)
	case "/game": // ?gameID=&boardID=&bingo
		h.getGame(w, r)
	case "/game/check_board": // ?gameID=&boardID=&type=
		h.checkBoard(w, r)
	case "/help":
		h.getHelp(w, r)
	case "/about":
		h.getAbout(w, r)
	default:
		http.NotFound(w, r)
	}
}

// servePost handles various POST requests.
func (h *httpsHandler) servePost(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/game":
		h.createGame(w, r)
	case "/game/draw_number": // ?gameID=
		h.drawNumber(w, r)
	case "/game/boards": // ?n=
		h.createBoards(w, r)
	default:
		http.NotFound(w, r)
	}
}

// httpError writes the message with statusCode to the response.
func httpError(w http.ResponseWriter, message string, statusCode int) {
	if len(message) == 0 {
		message = http.StatusText(statusCode)
	}
	http.Error(w, message, statusCode)
}

// getGames renders the games page onto the response with the game infos.
func (h httpsHandler) getGames(w http.ResponseWriter, r *http.Request) {
	handleGames(w, h.gameInfos)
}

// getGame renders the game page onto the response with the game of the 'gameID' query parameter.
// The 'boardID' and 'bingo' query parameters are also used to forward the results of a BINGO check.
func (h httpsHandler) getGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("gameID")
	boardID := r.URL.Query().Get("boardID")
	hasBingo := r.URL.Query().Has("bingo")
	var g *bingo.Game
	switch len(id) {
	case 0:
		g = new(bingo.Game)
	default:
		var ok bool
		g, ok = parseGame(id, w)
		if !ok {
			return
		}
	}
	handleGame(w, *g, boardID, hasBingo)
}

// hetHelp renders the help page onto the response.
func (httpsHandler) getHelp(w http.ResponseWriter, r *http.Request) {
	handleHelp(w)
}

// hetHelp renders the about page onto the response.
func (httpsHandler) getAbout(w http.ResponseWriter, r *http.Request) {
	handleAbout(w)
}

// checkBoard checks the board on the game with a checkType using the 'gameID', 'boardID', and 'type' query parameters.
// The results of the check are included as query parameters onto a redirect to the game page.'
func (httpsHandler) checkBoard(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")
	g, ok := parseGame(gameID, w)
	if !ok {
		return
	}
	boardID := r.URL.Query().Get("boardID")
	b, err := bingo.BoardFromID(boardID)
	if err != nil {
		message := fmt.Sprintf("getting board from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
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

// createGame redirects to a game that has not had any numbers drawn
func (h *httpsHandler) createGame(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

// drawNumber draws a new number for the game specified by the request's 'gameID' form parameter.
// The response is redirected to the updated game.  It's updated state is stored in the game infos slice.
func (h *httpsHandler) drawNumber(w http.ResponseWriter, r *http.Request) {
	gameID := r.FormValue("gameID")
	g, ok := parseGame(gameID, w)
	if !ok {
		return
	}
	before := g.NumbersLeft()
	g.DrawNumber()
	after := g.NumbersLeft()
	if before == after {
		http.Redirect(w, r, "/game?gameID="+gameID, http.StatusNotModified)
		return
	}
	id2, err := g.ID()
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
		ID:          id2,
		ModTime:     modTime,
		NumbersLeft: after,
	}
	h.gameInfos[0] = gi // set first
	http.Redirect(w, r, "/game?gameID="+gi.ID, http.StatusSeeOther)
}

// createBoards creates 'n' boards as specified by the request's form parameter, attaching the boards in a zip file.
func (h httpsHandler) createBoards(w http.ResponseWriter, r *http.Request) {
	nQueryParam := r.FormValue("n")
	n, err := strconv.Atoi(nQueryParam)
	if err != nil {
		message := fmt.Sprintf("%v: example: /game/boards?n=5 creates 5 unique boards", err)
		httpError(w, message, http.StatusBadRequest)
		return
	}
	if n < 1 || n > 1000 {
		httpError(w, "n must be be between 1 and 100", http.StatusBadRequest)
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
		if err := handleExportBoard(f, *board); err != nil {
			message := fmt.Sprintf("unexpected problam adding board %v to zip file: %v", i, err)
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

// withGzip wraps the handler with a handler that writes responses using gzip compression when accepted.
func withGzip(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}
		gzw := gzip.NewWriter(w)
		defer gzw.Close()
		wrw := wrappedResponseWriter{
			Writer:         gzw,
			ResponseWriter: w,
		}
		wrw.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(wrw, r)
	}
}

// wrappedResponseWriter wraps response writing with another writer.
type wrappedResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// Write delegates the write to the wrapped writer.
func (wrw wrappedResponseWriter) Write(p []byte) (n int, err error) {
	return wrw.Writer.Write(p)
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
