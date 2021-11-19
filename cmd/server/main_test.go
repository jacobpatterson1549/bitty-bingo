package main

import (
	"bytes"
	"flag"
	"reflect"
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
