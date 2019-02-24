//go:generate go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 1024 --host 127.0.0.1,::1,localhost --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestRunServer_TLS(t *testing.T) {
	testKeysAreAvailable(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := newPort()

	done := make(chan error)
	go func() {
		done <- RunServer(
			ctx,
			config.ServiceConfig{
				Port: port,
				TLS: &config.TLS{
					PublicKey:  "cert.pem",
					PrivateKey: "key.pem",
				},
			},
			http.HandlerFunc(dummyHandler),
		)
	}()

	client, err := httpsClient("cert.pem")
	if err != nil {
		t.Error(err)
		return
	}

	<-time.After(100 * time.Millisecond)

	resp, err := client.Get(fmt.Sprintf("https://localhost:%d", port))
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

func TestRunServer_plain(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := newPort()

	done := make(chan error)
	go func() {
		done <- RunServer(
			ctx,
			config.ServiceConfig{Port: port},
			http.HandlerFunc(dummyHandler),
		)
	}()

	<-time.After(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
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

func TestRunServer_disabledTLS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error)

	port := newPort()

	go func() {
		done <- RunServer(
			ctx,
			config.ServiceConfig{
				Port: port,
				TLS: &config.TLS{
					IsDisabled: true,
				}},
			http.HandlerFunc(dummyHandler),
		)
	}()

	<-time.After(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
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

func TestRunServer_errBadKeys(t *testing.T) {
	done := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		done <- RunServer(
			ctx,
			config.ServiceConfig{TLS: &config.TLS{
				PublicKey:  "unknown",
				PrivateKey: "unknown",
			}},
			http.HandlerFunc(dummyHandler),
		)
	}()
	if err := <-done; err == nil || err.Error() != "open unknown: no such file or directory" {
		t.Error(err)
	}
}

func TestInitHTTPDefaultTransport(t *testing.T) {
	cleanUp()
	defer cleanUp()

	InitHTTPDefaultTransport(config.ServiceConfig{
		DialerTimeout:         time.Hour,
		DialerKeepAlive:       2 * time.Hour,
		DialerFallbackDelay:   3 * time.Hour,
		DisableCompression:    true,
		DisableKeepAlives:     true,
		MaxIdleConns:          123,
		MaxIdleConnsPerHost:   234,
		IdleConnTimeout:       time.Second,
		ResponseHeaderTimeout: 2 * time.Second,
		ExpectContinueTimeout: 3 * time.Minute,
	})

	transport := http.DefaultTransport.(*http.Transport)

	if err := equalDuration("IdleConnTimeout", transport.IdleConnTimeout, time.Second); err != nil {
		t.Error(err)
	}

	if err := equalDuration("ResponseHeaderTimeout", transport.ResponseHeaderTimeout, 2*time.Second); err != nil {
		t.Error(err)
	}

	if err := equalDuration("ExpectContinueTimeout", transport.ExpectContinueTimeout, 3*time.Minute); err != nil {
		t.Error(err)
	}

	if err := equalInt("MaxIdleConns", transport.MaxIdleConns, 123); err != nil {
		t.Error(err)
	}

	if err := equalInt("MaxIdleConnsPerHost", transport.MaxIdleConnsPerHost, 234); err != nil {
		t.Error(err)
	}

	if !transport.DisableCompression {
		t.Error("the DisableCompression value is 'false' when it should be 'true'")
	}

	if !transport.DisableKeepAlives {
		t.Error("the DisableKeepAlives value is 'false' when it should be 'true'")
	}
}

func equalDuration(name string, actual, expected time.Duration) error {
	if actual != expected {
		return fmt.Errorf("unexpected %s. have: %v, want: %v", name, actual, expected)
	}
	return nil
}

func equalInt(name string, actual, expected int) error {
	if actual != expected {
		return fmt.Errorf("unexpected %s. have: %d, want: %d", name, actual, expected)
	}
	return nil
}

func cleanUp() {
	onceTransportConfig = sync.Once{}
}

func dummyHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, %q", html.EscapeString(req.URL.Path))
}

func testKeysAreAvailable(t *testing.T) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	for _, k := range []string{"cert.pem", "key.pem"} {
		var exists bool
		for _, file := range files {
			if file.Name() == k {
				exists = true
				break
			}
		}
		if !exists {
			t.Errorf("file %s not present", k)
		}
	}
}

func httpsClient(cert string) (*http.Client, error) {
	cer, err := ioutil.ReadFile(cert)
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(cer)
	if !ok {
		return nil, errors.New("failed to parse root certificate")
	}
	tlsConf := &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		},
		RootCAs: roots,
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}}, nil
}

// newPort returns random port numbers to avoid port collisions during the tests
func newPort() int {
	return 16666 + rand.Intn(40000)
}
