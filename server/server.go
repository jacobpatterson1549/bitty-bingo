// Package server provides a restful TCP interface to manage bingo games and boards.
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jacobpatterson1549/bitty-bingo/server/handler"
)

// Config is used to create a server.
type Config struct {
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

const (
	// readDur is the maximum time taken to read a request before timing out.
	readDur = 60 * time.Second
	// writeDur is the maximum time taken to write a request before timing out.
	writeDur = 60 * time.Second
	// stopDur is the maximum time allowed for the server to shut down.
	stopDur = 5 * time.Second
)

// Server manages bingo games and creates boards.
type Server struct {
	Config
	httpsServer *http.Server
	httpServer  *http.Server
}

// NewServer initializes HTTP and HTTPS tcp servers from the Configs
func (cfg Config) NewServer() (*Server, error) {
	httpsHandler, err := cfg.httpsHandler()
	if err != nil {
		return nil, fmt.Errorf("creating httpsHandler: %v", err)
	}
	httpHandler := cfg.httpHandler()
	s := Server{
		Config:      cfg,
		httpsServer: httpServer(cfg.HTTPSPort, httpsHandler),
		httpServer:  httpServer(cfg.HTTPPort, httpHandler),
	}
	return &s, nil
}

// Run starts the HTTP and HTTPS TCP servers.
func (s *Server) Run() <-chan error {
	errC := make(chan error, 2)
	go s.serveTCP(s.httpsServer, "https", errC, true)
	if s.HTTPSRedirect {
		go s.serveTCP(s.httpServer, "http", errC, false)
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
	switch {
	case err1 != nil:
		return err2
	case err2 != nil:
		return err2
	}
	return nil
}

// httpServer creates a http server on the port with the handler, using default read and write timeouts.
func httpServer(port string, h http.Handler) *http.Server {
	s := http.Server{
		Addr:         ":" + port,
		Handler:      h,
		ReadTimeout:  readDur,
		WriteTimeout: writeDur,
	}
	return &s
}

// serveTCP servers the server on TCP for the address.
// TLS certificates are loaded if the server is HTTPS and has a separarte server that redirects HTTP requests to HTTPS.
func (cfg Config) serveTCP(svr *http.Server, name string, errC chan<- error, https bool) {
	l, err := net.Listen("tcp", svr.Addr)
	if err != nil {
		errC <- fmt.Errorf("listening to tcp at address %q: %v", svr.Addr, err)
		return
	}
	defer l.Close()
	if https && cfg.HTTPSRedirect {
		certificate, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			errC <- fmt.Errorf("loading TLS certificates (%q and %q): %v", cfg.TLSCertFile, cfg.TLSKeyFile, err)
			return
		}
		cfg := tls.Config{
			NextProtos:   []string{"http/1.1"},
			Certificates: []tls.Certificate{certificate},
			MinVersion:   tls.VersionTLS13,
		}
		l = tls.NewListener(l, &cfg)
	}
	svr.Serve(l) // BLOCKING
}

// httpHandler creates a HTTP handler that redirects all requests to HTTPS.
// Responses are returned gzip compression when allowed.
func (cfg Config) httpHandler() http.Handler {
	h := handler.Redirect(cfg.HTTPSPort)
	return handler.WithGzip(h)
}

// httpsHandler creates a HTTP handler to serve the site.
// The gameCount and time function are validated used from the config in the handler
// Responses are returned gzip compression when allowed.
func (cfg Config) httpsHandler() (http.Handler, error) {
	h, err := handler.Handler(cfg.GameCount, cfg.Time)
	if err != nil {
		return nil, fmt.Errorf("creating root handler for server: %v", err)
	}
	return handler.WithGzip(h), nil
}
