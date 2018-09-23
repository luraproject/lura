package http

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"testing"

	"github.com/devopsfaith/krakend/config"
)

func TestRunServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)
	go func() {
		done <- RunServer(
			ctx,
			config.ServiceConfig{Port: 9999},
			http.HandlerFunc(dummyHandler),
		)
	}()

	resp, err := http.Get("http://localhost:9999")
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}
	cancel()

	if err = <-done; err != nil {
		t.Error(err)
	}
}

func TestRunServer_err(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error)
	for _, tc := range []struct {
		cfg *config.TLS
		err error
	}{
		{
			cfg: &config.TLS{},
			err: ErrPublicKey,
		},
		{
			cfg: &config.TLS{
				PublicKey: "unknown",
			},
			err: ErrPrivateKey,
		},
	} {
		go func() {
			done <- RunServer(
				ctx,
				config.ServiceConfig{TLS: tc.cfg},
				http.HandlerFunc(dummyHandler),
			)
		}()
		if err := <-done; err != tc.err {
			t.Error(err)
		}
	}
}

func dummyHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, %q", html.EscapeString(req.URL.Path))
}
