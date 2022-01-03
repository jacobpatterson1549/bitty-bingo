package handler

import "net/http"

// HTTPSRedirectPort is a handler that redirects requests to HTTPS uris.
type HTTPSRedirectPort string

// ServeHTTP redirects requests to HTTPS on the port if it is not standard (443).
func (port HTTPSRedirectPort) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	httpsURI := "https://" + r.URL.Hostname()
	if len(r.URL.Port()) != 0 && port != "443" {
		httpsURI += ":" + string(port)
	}
	httpsURI += r.URL.Path
	http.Redirect(w, r, httpsURI, http.StatusMovedPermanently)
}
