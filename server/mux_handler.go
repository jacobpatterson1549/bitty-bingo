package server

import "net/http"

// MuxHandler is http Handler that is a map of methods to path handler maps
type MuxHandler map[string]map[string]http.HandlerFunc

// ServeHTTP serves to the path for the method of the request on the handler if such a Handler exists.
func (m MuxHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	methodHandlers, ok := m[r.Method]
	if !ok {
		httpError(w, "", http.StatusMethodNotAllowed)
		return
	}
	h, ok := methodHandlers[r.URL.Path]
	if !ok {
		httpError(w, "", http.StatusNotFound)
		return
	}
	h.ServeHTTP(w, r)
}
