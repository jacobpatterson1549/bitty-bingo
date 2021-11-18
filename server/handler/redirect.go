package handler

import "net/http"

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
