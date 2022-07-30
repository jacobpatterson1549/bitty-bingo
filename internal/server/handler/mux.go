package handler

import "net/http"

// Mux is http Handler that maps methods to paths to handlers.
type Mux map[string]map[string]http.HandlerFunc

// ServeHTTP serves to the path for the method of the request on the handler if such a Handler exists.
func (m Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	methodHandlers, ok := m[r.Method]
	if !ok {
		httpError(w, http.StatusMethodNotAllowed)
		return
	}
	h, ok := methodHandlers[r.URL.Path]
	if !ok {
		httpError(w, http.StatusNotFound)
		return
	}
	h.ServeHTTP(w, r)
}

// httpError writes the message for the statusCode to the response.
func httpError(w http.ResponseWriter, statusCode int) {
	message := http.StatusText(statusCode)
	http.Error(w, message, statusCode)
}
