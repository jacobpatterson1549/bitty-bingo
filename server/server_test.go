package server

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		wantOk bool
	}{
		{
			name: "zero-value config [bad games count, time func]",
		},
		{
			name: "happy path",
			cfg: Config{
				GameCount: 100,
				Time: func() string {
					return "time"
				},
				HTTPPort:      "a",
				HTTPSPort:     "b",
				TLSCertFile:   "c",
				TLSKeyFile:    "d",
				HTTPSRedirect: true,
			},
			wantOk: true,
		},
	}
	for i, test := range tests {
		s, err := test.cfg.NewServer()
		switch {
		case !test.wantOk:
			if err == nil {
				t.Errorf("test %v (%v): wanted error", i, test.name)
			}
		case err != nil:
			t.Errorf("test %v (%v): unwanted error: %v", i, test.name, err)
		case s.httpsServer == nil:
			t.Errorf("test %v (%v): HTTPS server not set", i, test.name)
		case s.httpServer == nil:
			t.Errorf("test %v (%v): HTTP server not set", i, test.name)
		default:
			s.Config.Time = nil // functions are not comparable
			test.cfg.Time = nil // "
			if want, got := test.cfg, s.Config; !reflect.DeepEqual(want, got) {
				t.Errorf("test %v (%v): config not copied to server:\nwanted: %v\ngot:    %v", i, test.name, want, got)
			}
		}
	}

}

func TestServerShutdown(t *testing.T) {
	s := Server{
		httpsServer: new(http.Server),
		httpServer:  new(http.Server),
	}
	ctx := context.Background()
	if err := s.Shutdown(ctx); err != nil {
		t.Errorf("unwanted error shutting down server: %v", err)
	}
}
