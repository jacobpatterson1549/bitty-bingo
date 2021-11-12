package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

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
	readDur  = 60 * time.Second
	writeDur = 60 * time.Second
	stopDur  = 5 * time.Second
)

type Server struct {
	Config
	httpsServer *http.Server
	httpServer  *http.Server
}

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

// Shutdown
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

func httpServer(port string, h http.Handler) *http.Server {
	s := http.Server{
		Addr:         ":" + port,
		Handler:      h,
		ReadTimeout:  readDur,
		WriteTimeout: writeDur,
	}
	return &s
}

func (cfg Config) serveTCP(svr *http.Server, name string, errC chan<- error, https bool) {
	l, err := net.Listen("tcp", svr.Addr)
	if err != nil {
		errC <- err
		return
	}
	defer l.Close()
	if cfg.HTTPSRedirect {
		certificate, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			errC <- fmt.Errorf("loading TLS certificates: %v", err)
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
