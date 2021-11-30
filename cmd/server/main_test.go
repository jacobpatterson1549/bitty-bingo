package main

import (
	"bytes"
	"flag"
	"log"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/jacobpatterson1549/bitty-bingo/server"
)

func TestFlagSet(t *testing.T) {
	t.Run("name", func(t *testing.T) {
		var cfg server.Config
		wantName := "program_name_325"
		fs := flagSet(&cfg, wantName)
		if gotName := fs.Name(); wantName != gotName {
			t.Errorf("names not equal:\nwanted: %q\ngot:    %q", wantName, gotName)
		}
	})
	t.Run("help", func(t *testing.T) {
		var cfg server.Config
		var buf bytes.Buffer
		fs := flagSet(&cfg, "name1")
		fs.SetOutput(&buf)
		fs.Init("name2", flag.ContinueOnError) // don't actually exit
		err := fs.Parse([]string{"-h"})
		switch {
		case err != flag.ErrHelp:
			t.Errorf("wanted help error")
		case buf.Len() == 0:
			t.Errorf("wanted help to be printed")
		}
	})
}

func TestParseServerConfig(t *testing.T) {
	for i, test := range parseServerConfigTests {
		var cfg server.Config
		fs := flagSet(&cfg, "")
		parseServerConfig(&cfg, fs, test.programArgs, test.portOverride, test.hasPortOverride)
		if cfg.Time == nil || len(cfg.Time()) == 0 {
			t.Errorf("test %v (%v): time func not set or returns nothing", i, test.name)
		}
		cfg.Time = nil // funcs are not comparable
		if want, got := test.wantConfig, cfg; !reflect.DeepEqual(want, got) {
			t.Errorf("test %v (%v): configs are not equal:\nwanted: %#v\ngot:    %#v", i, test.name, want, got)
		}
	}
}

func TestRunServer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that runs an HTTPS server on a tcp port")
	}
	okConfig := server.Config{
		GameCount: 10,
		Time:      func() string { return "time" },
	}
	tests := []struct {
		name          string
		serverConfig  server.Config
		httpsRedirect bool
		wantLogPart   string
		wantErrPart   string
	}{
		{
			name:        "bad config",
			wantErrPart: "problem creating server",
		},
		{
			name:          "server with redirect",
			serverConfig:  okConfig,
			httpsRedirect: true,
			wantLogPart:   "https://",
			wantErrPart:   "already in use",
		},
		{
			name:          "server with PORT specified",
			serverConfig:  okConfig,
			httpsRedirect: false,
			wantLogPart:   "http://",
			wantErrPart:   "already in use",
		},
	}
	for i, test := range tests {
		svr := httptest.NewServer(nil)
		port := serverPort(t, svr)
		cfg := test.serverConfig
		cfg.HTTPSPort = port
		cfg.HTTPSRedirect = test.httpsRedirect
		var buf bytes.Buffer
		log := log.New(&buf, "", 0)
		gotErr := runServer(cfg, log)
		switch {
		case gotErr == nil:
			t.Errorf("test %v (%v): wanted error running server on same port as test server", i, test.name)
		case !strings.Contains(buf.String(), test.wantLogPart):
			t.Errorf("test %v (%v): wanted log to notify that server started on %q, got %q", i, test.name, test.wantLogPart, buf.String())
		case !strings.Contains(gotErr.Error(), test.wantErrPart):
			t.Errorf("test %v (%v): wanted address in use error, got %v", i, test.name, gotErr)
		}
		svr.Close()
	}
}

func serverPort(t *testing.T, svr *httptest.Server) string {
	t.Helper()
	u, err := url.Parse(svr.URL)
	if err != nil {
		t.Fatalf("getting test server port: %v", err)
	}
	return u.Port()
}

var (
	sampleProgramArgs = []string{
		"--http-port=8001",
		"--https-port=8000",
		"--tls-cert-file=/home/jacobpatterson1549/tls-cert.pem",
		"--tls-key-file=/home/jacobpatterson1549/tls-key.pem",
		"--game-count=33",
	}
	parseServerConfigTests = []struct {
		name            string
		programArgs     []string
		wantConfig      server.Config
		portOverride    string
		hasPortOverride bool
	}{
		{
			name: "no args (use defaults)",
			wantConfig: server.Config{
				HTTPSRedirect: true,
				HTTPPort:      "80",
				HTTPSPort:     "443",
				GameCount:     10,
			},
		},
		{
			name:        "all flags",
			programArgs: sampleProgramArgs,
			wantConfig: server.Config{
				GameCount:     33,
				HTTPPort:      "8001",
				HTTPSPort:     "8000",
				TLSCertFile:   "/home/jacobpatterson1549/tls-cert.pem",
				TLSKeyFile:    "/home/jacobpatterson1549/tls-key.pem",
				HTTPSRedirect: true,
			},
		},
		{
			name:        "PORT should override HTTPS port and not redirect",
			programArgs: sampleProgramArgs,
			wantConfig: server.Config{
				GameCount:     33,
				HTTPPort:      "8001",
				HTTPSPort:     "444",
				TLSCertFile:   "/home/jacobpatterson1549/tls-cert.pem",
				TLSKeyFile:    "/home/jacobpatterson1549/tls-key.pem",
				HTTPSRedirect: false,
			},
			portOverride:    "444",
			hasPortOverride: true,
		},
	}
)
