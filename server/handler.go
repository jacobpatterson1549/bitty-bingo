package server

import (
	"archive/zip"
	"compress/gzip"
	"encoding/json"
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
	}
	return withGzip(http.HandlerFunc(h.serveHTTPS))
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
}

func (h *httpsHandler) serveHTTPS(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		h.serveGet(w, r)
	case "POST":
		h.servePost(w, r)
	default:
		httpError(w, "", http.StatusMethodNotAllowed)
	}
}

func (h *httpsHandler) servePost(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/game/create":
		h.createGame(w, r)
	case "/game/boards": // ?n=
		h.createBoards(w, r)
	case "/game/check_board": // ?game=&board=&check=
		h.checkBoard(w, r)
	case "/game/draw_number": // ?game=
		h.drawNumber(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h httpsHandler) serveGet(w http.ResponseWriter, r *http.Request) {
	// TODO: create parent html wrapper page with nav bar(games list, help, about links)
	switch r.URL.Path {
	case "/":
		h.getGames(w, r)
	case "/game":
		h.getGame(w, r)
	case "/help":
		h.getHelp(w, r)
	case "/about":
		h.getAbout(w, r)
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

func (h *httpsHandler) createGame(w http.ResponseWriter, r *http.Request) {
	var g bingo.Game
	if err := handleGame(w, g); err != nil {
		message := fmt.Sprintf("rendering new game: %v", err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
}

func (h httpsHandler) getGames(w http.ResponseWriter, r *http.Request) {
	if err := handleGames(w, h.gameInfos); err != nil {
		message := fmt.Sprintf("rendering games list: %v", err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
}

func (h httpsHandler) getGame(w http.ResponseWriter, r *http.Request) {
	gameQueryParam := r.URL.Query().Get("id")
	var g bingo.Game
	if err := json.Unmarshal([]byte(gameQueryParam), &g); err != nil {
		message := fmt.Sprintf("getting game from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
	}
	if err := handleGame(w, g); err != nil {
		message := fmt.Sprintf("rendering game: %v", err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
}

func (httpsHandler) getHelp(w http.ResponseWriter, r *http.Request) {
	if err := handleHelp(w); err != nil {
		message := fmt.Sprintf("rendering help page: %v", err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
}

func (httpsHandler) getAbout(w http.ResponseWriter, r *http.Request) {
	if err := handleAbout(w); err != nil {
		message := fmt.Sprintf("rendering about page: %v", err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
}

func (httpsHandler) checkBoard(w http.ResponseWriter, r *http.Request) {
	gameQueryParam := r.URL.Query().Get("gameID")
	var g bingo.Game
	if err := json.Unmarshal([]byte(gameQueryParam), &g); err != nil {
		message := fmt.Sprintf("getting game from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
	}
	boardQueryParam := r.URL.Query().Get("board")
	var b bingo.Board
	if err := json.Unmarshal([]byte(boardQueryParam), &b); err != nil {
		message := fmt.Sprintf("getting board from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
	}
	checkType := r.URL.Query().Get("type")
	// var result bool
	switch checkType {
	case "HasLine":
		// result = b.HasLine(g)
	case "IsFilled":
		// result = b.IsFilled(g)
	default:
		message := fmt.Sprintf("unknown checkType %q", checkType)
		httpError(w, message, http.StatusBadRequest)
	}
	// fmt.Fprint(w, result)
	// TODO: redirect to /game with query of board and result
}

func (h *httpsHandler) drawNumber(w http.ResponseWriter, r *http.Request) {
	gameQueryParam := r.Form.Get("game")
	var g bingo.Game
	if err := json.Unmarshal([]byte(gameQueryParam), &g); err != nil {
		message := fmt.Sprintf("getting game from query parameter: %v", err)
		httpError(w, message, http.StatusBadRequest)
		return
	}
	before := g.NumbersLeft()
	g.DrawNumber()
	after := g.NumbersLeft()
	if before != after {
		if len(h.gameInfos) < cap(h.gameInfos) {
			h.gameInfos = append(h.gameInfos, gameInfo{}) // increase length
		}
		copy(h.gameInfos[1:], h.gameInfos) // shift right
		var gi gameInfo
		// TODO: create game info for UTC
		h.gameInfos[0] = gi // set first
	}
	// TODO: redirect to updated game
}

func (h httpsHandler) createBoards(w http.ResponseWriter, r *http.Request) {
	nQueryParam := r.Form.Get("n")
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
			message := fmt.Sprintf("problem creating zip file #%v: %v", i, fileName)
			httpError(w, message, http.StatusInternalServerError)
			return
		}
		board := bingo.NewBoard() // TODO: ensure boards are unique
		if err := handleExportBoard(f, *board); err != nil {
			message := fmt.Sprintf("problam adding board %v to zip file: %v", i, err)
			httpError(w, message, http.StatusInternalServerError)
			return
		}
	}
	if err := z.Close(); err != nil {
		message := fmt.Sprintf("problem closing zip file: %v", err)
		httpError(w, message, http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=bingo-boards.zip")
	// TODO: redirect to games list
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
