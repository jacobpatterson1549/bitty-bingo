package server

import (
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/jacobpatterson1549/bitty-bingo/bingo"
)

func (cfg Config) httpHandler() http.Handler {
	return withGzip(http.HandlerFunc(httpsRedirectHandler(cfg.HTTPSPort)))
}

func (cfg Config) httpsHandler() http.Handler {
	h := httpsHandler{
		gameInfos: make([]gameInfo, 0, cfg.GameCount),
		time:      cfg.Time,
	}
	return withGzip(&h)
}

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

type gameInfo struct {
	ID          string
	ModTime     string
	NumbersLeft int
}

type httpsHandler struct {
	gameInfos []gameInfo
	time      func() string
}

func (h *httpsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.serveGet(w, r)
	case "POST":
		h.servePost(w, r)
	default:
		httpError(w, "", http.StatusMethodNotAllowed)
	}
}

func (h httpsHandler) serveGet(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.getGames(w, r)
	case "/game": // ?gameID=
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

func (h *httpsHandler) servePost(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/game/create":
		h.createGame(w, r)
	case "/game/draw_number": // ?gameID=
		h.drawNumber(w, r)
	case "/game/boards": // ?n=
		h.createBoards(w, r)
	default:
		http.NotFound(w, r)
	}
}

func httpError(w http.ResponseWriter, message string, statusCode int) {
	if len(message) == 0 {
		message = http.StatusText(statusCode)
	}
	http.Error(w, message, statusCode)
}

func (h httpsHandler) getGames(w http.ResponseWriter, r *http.Request) {
	handleGames(w, h.gameInfos)
}

func (h httpsHandler) getGame(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("gameID")
	var g *bingo.Game
	switch len(id) {
	case 0:
		g = new(bingo.Game)
	default:
		var err error
		g, err = bingo.GameFromID(id)
		if err != nil {
			message := fmt.Sprintf("getting game from query parameter: %v", err)
			httpError(w, message, http.StatusBadRequest)
			return
		}
	}
	handleGame(w, *g)
}

func (httpsHandler) getHelp(w http.ResponseWriter, r *http.Request) {
	handleHelp(w)
}

func (httpsHandler) getAbout(w http.ResponseWriter, r *http.Request) {
	handleAbout(w)
}

func (httpsHandler) checkBoard(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameID")
	g, err := bingo.GameFromID(gameID)
	if err != nil {
		message := fmt.Sprintf("getting game from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
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
	fmt.Fprint(w, result)
	// TODO: redirect to /game with query of board and result
}

func (h *httpsHandler) createGame(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/game", http.StatusSeeOther)
}

func (h *httpsHandler) drawNumber(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("gameID")
	g, err := bingo.GameFromID(id)
	if err != nil {
		message := fmt.Sprintf("getting game from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
		return
	}
	before := g.NumbersLeft()
	g.DrawNumber()
	after := g.NumbersLeft()
	if before == after {
		http.Redirect(w, r, "/game?gameID="+id, http.StatusNotModified)
		return
	}
	id2, err := g.ID()
	if err != nil {
		message := fmt.Sprintf("unexpected problem getting id after drawing tile from game with a VALID id %q: %v", id, err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
	if len(h.gameInfos) < cap(h.gameInfos) {
		h.gameInfos = append(h.gameInfos, gameInfo{}) // increase length
	}
	copy(h.gameInfos[1:], h.gameInfos) // shift right
	modTime := h.time()
	gi := gameInfo{
		ID:          id2,
		ModTime:     modTime,
		NumbersLeft: after,
	}
	h.gameInfos[0] = gi // set first
	http.Redirect(w, r, "/game?gameID="+gi.ID, http.StatusSeeOther)
}

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
	z := zip.NewWriter(w)
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
	w.Header().Add("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=bingo-boards.zip")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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
