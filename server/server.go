package server

import (
	"context"
	"crypto/tls"
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
}

const (
	readDur  = 60 * time.Second
	writeDur = 60 * time.Second
	stopDur  = 5 * time.Second
)

type Server struct {
	Config
	httpServer  *http.Server
	httpsServer *http.Server
}

func (cfg Config) NewServer() *Server {
	s := Server{
		Config:      cfg,
		httpServer:  httpServer(cfg.HTTPPort, cfg.httpHandler()),
		httpsServer: httpServer(cfg.HTTPSPort, cfg.httpsHandler()),
	}
	return &s
}

// Run starts the HTTP and HTTPS TCP servers.
func (s *Server) Run() <-chan error {
	errC := make(chan error, 2)
	if s.HTTPSRedirect {
		go s.serveTCP(s.httpServer, "http", errC, false)
	}
	go s.serveTCP(s.httpsServer, "https", errC, true)
	return errC
}

// Shutdown
func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancelFunc := context.WithTimeout(ctx, stopDur)
	defer cancelFunc()
	var err1, err2 error
	if s.httpServer != nil {
		err1 = s.httpServer.Shutdown(ctx)
	}
	if s.httpsServer != nil {
		err2 = s.httpsServer.Shutdown(ctx)
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
			errC <- err
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
