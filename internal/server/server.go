// Package server provides a restful TCP interface to manage bingo games and boards.
package server

import (
	"context"
	"crypto/tls"
	"image"
	"net/http"
	"time"

	"github.com/jacobpatterson1549/bitty-bingo/internal/server/handler"
	"github.com/jacobpatterson1549/bitty-bingo/internal/server/handler/barcode"
)

type (
	// Server manages bingo games and creates boards.
	Server struct {
		config      Config
		httpsServer *http.Server
		httpServer  *http.Server
	}
	// Config is used to create a server.
	Config struct {
		// HTTPSRedirect causes the server to run a redirect from HTTPPORT to HTTPSPort when set
		HTTPSRedirect bool
		// HTTPPort is the HTTP port the server runs on.
		HTTPPort string
		// HTTPSPort is the HTTPS port the server runs on.
		HTTPSPort string
		// TLSCertFile is the public HTTPS TLS certificate file name.
		TLSCertFile string
		// TLSKeyFile is the private HTTPS TLS key file name.
		TLSKeyFile string
		// GameCount is the number of game states kept in the game list.
		GameCount int
		// Time is a function that can add a timestamp to parts of the site.
		Time func() string
	}
)

const (
	// readDur is the maximum time taken to read a request before timing out.
	readDur = 60 * time.Second
	// writeDur is the maximum time taken to write a request before timing out.
	writeDur = 60 * time.Second
	// stopDur is the maximum time allowed for the server to shut down.
	stopDur = 5 * time.Second
)

// NewServer initializes HTTP and HTTPS TCP servers from the Configs
func (cfg Config) NewServer() *Server {
	httpsHandler := cfg.httpsHandler()
	httpHandler := cfg.httpHandler()
	s := Server{
		config:      cfg,
		httpsServer: httpServer(cfg.HTTPSPort, httpsHandler, true),
		httpServer:  httpServer(cfg.HTTPPort, httpHandler, false),
	}
	return &s
}

// Run starts the HTTP and HTTPS TCP servers.
func (s *Server) Run() <-chan error {
	errC := make(chan error, 2)
	go s.listenAndServe(s.httpsServer, errC, true)
	if s.config.HTTPSRedirect {
		go s.listenAndServe(s.httpServer, errC, false)
	}
	return errC
}

// Shutdown waits for the HTTP and HTTPS servers to shut down.
func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancelFunc := context.WithTimeout(ctx, stopDur)
	defer cancelFunc()
	var err1, err2 error
	if s.httpsServer != nil {
		err1 = s.httpsServer.Shutdown(ctx)
	}
	if s.httpServer != nil {
		err2 = s.httpServer.Shutdown(ctx)
	}
	return firstNonNilError(err1, err2)
}

// httpServer creates a http server on the port with the handler, using default read and write timeouts.
func httpServer(port string, h http.Handler, https bool) *http.Server {
	svr := http.Server{
		Addr:         ":" + port,
		Handler:      h,
		ReadTimeout:  readDur,
		WriteTimeout: writeDur,
	}
	if https {
		svr.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
		}
	}
	return &svr
}

// listenAndServe servers the server on TCP for the address.
// TLS certificates are loaded if the server is HTTPS and has a separate server that redirects HTTP requests to HTTPS.
func (s Server) listenAndServe(svr *http.Server, errC chan<- error, https bool) {
	switch {
	case https && s.config.HTTPSRedirect:
		errC <- svr.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile) // BLOCKING
	default:
		errC <- svr.ListenAndServe() // BLOCKING
	}
}

// httpHandler creates a HTTP handler that redirects all requests to HTTPS.
// Responses are returned gzip compression when allowed.
func (cfg Config) httpHandler() http.Handler {
	h := handler.HTTPSRedirectPort(cfg.HTTPSPort)
	return handler.WithGzip(h)
}

// httpsHandler creates a HTTP handler to serve the site.
// The gameCount and time function are validated used from the config in the handler
// Responses are returned gzip compression when allowed.
func (cfg Config) httpsHandler() http.Handler {
	h := handler.Handler(cfg.GameCount, cfg.Time, cfg)
	return handler.WithGzip(h)
}

func (c Config) BarCode(format string, text string, width, height int) (image.Image, error) {
	f := c.barCodeFormat(format)
	return barcode.Image(f, text, width, height)
}

func (c Config) barCodeFormat(f string) barcode.Format {
	switch f {
	case "aztec":
		return barcode.AZTEC
	case "data_matrix":
		return barcode.DATA_MATRIX
	default:
		return barcode.QR_CODE
	}
}

// firstNonNill returns the first error that is not nil
func firstNonNilError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
