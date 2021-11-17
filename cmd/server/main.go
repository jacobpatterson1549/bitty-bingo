// Package main runs a bingo server on the command line..
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jacobpatterson1549/bitty-bingo/server"
)

// main runs a bingo server
func main() {
	programName, programArgs := os.Args[0], os.Args[1:]
	portOverride, hasPortOverride := os.LookupEnv("PORT")
	var cfg server.Config
	fs := flagSet(&cfg, programName)
	parseServerConfig(&cfg, fs, programArgs, portOverride, hasPortOverride)
	runServer(cfg) // BLOCKING
}

// flagSet creates a flag set that sets the config
func flagSet(cfg *server.Config, programName string) *flag.FlagSet {
	fs := flag.NewFlagSet(programName, flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Runs the server")
		fmt.Fprintln(fs.Output(), "Providing PORT environment variable overrides the command line argument, runs the HTTPS Server on the specified port, does not provide a HTTP redirect, and does not load TLS certificates.")
		fs.PrintDefaults()
	}
	fs.StringVar(&cfg.HTTPPort, "http-port", "80", "The TCP port for HTTP requests.")
	fs.StringVar(&cfg.HTTPSPort, "https-port", "443", "The TCP port for HTTPS requests.")
	fs.StringVar(&cfg.TLSCertFile, "tls-cert-file", "", "The name of the TLS public certificate file")
	fs.StringVar(&cfg.TLSKeyFile, "tls-key-file", "", "The name of the TLS private key file")
	fs.IntVar(&cfg.GameCount, "game-count", 10, "The number of game states to keep in the history")
	return fs
}

// parseServerConfig parses command line flag set and environment variables into a server config
func parseServerConfig(cfg *server.Config, fs *flag.FlagSet, programArgs []string, portOverride string, hasPortOverride bool) {
	fs.Parse(programArgs)
	cfg.HTTPSRedirect = !hasPortOverride
	if hasPortOverride {
		cfg.HTTPSPort = portOverride
	}
	cfg.Time = func() string {
		return time.Now().UTC().String()
	}
}

// runServer creates and runs a bingo server from the config
func runServer(cfg server.Config) {
	s, err := cfg.NewServer()
	if err != nil {
		log.Fatalf("creating server: %v", err)
		return
	}
	done := make(chan os.Signal, 2)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	errC := s.Run()
	switch {
	case cfg.HTTPSRedirect:
		log.Printf("started server at https://127.0.0.1:%v", cfg.HTTPSPort)
	default:
		log.Printf("started server at http://127.0.0.1:%v", cfg.HTTPSPort)
	}
	select { // BLOCKING
	case err := <-errC:
		if err != nil {
			log.Printf("running server: %v", err)
		}
	case signal := <-done:
		log.Printf("handled signal: %v", signal)
	}
	ctx := context.Background()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalf("stopping server: %v", err)
		return
	}
	log.Printf("server stopped successfully")
}
