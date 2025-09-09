// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"html"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"golang.org/x/net/http2"
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
					CaCerts:    []string{"ca.pem"},
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

	// now lets initialize the global default transport and use a regular
	// client to connect to the server
	InitHTTPDefaultTransport(config.ServiceConfig{
		ClientTLS: &config.ClientTLS{
			CaCerts:             []string{"ca.pem"},
			DisableSystemCaPool: true,
		},
	})
	rawClient := http.Client{}
	resp, err = rawClient.Get(fmt.Sprintf("https://localhost:%d", port))
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

func TestRunServer_MTLS(t *testing.T) {
	testKeysAreAvailable(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := 36517
	done := make(chan error)

	cfg := config.ServiceConfig{
		Port: port,
		TLS: &config.TLS{
			Keys: []config.TLSKeyPair{
				{
					PublicKey:  "cert.pem",
					PrivateKey: "key.pem",
				},
			},
			CaCerts:    []string{"ca.pem"},
			EnableMTLS: true,
		},
		ClientTLS: &config.ClientTLS{
			AllowInsecureConnections: false, // we do not check the server cert
			CaCerts:                  []string{"ca.pem"},
			ClientCerts: []config.ClientTLSCert{
				{
					Certificate: "cert.pem",
					PrivateKey:  "key.pem",
				},
			},
		},
	}
	go func() {
		done <- RunServer(ctx, cfg, http.HandlerFunc(dummyHandler))
	}()

	client, err := mtlsClient("cert.pem", "key.pem")
	if err != nil {
		t.Error(err)
		return
	}

	<-time.After(1000 * time.Millisecond)

	resp, err := client.Get(fmt.Sprintf("https://localhost:%d", port))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	logger := logging.NoOp
	// since test are run in a suite, and `InitHTTPDefaultTransportWithLogger` is
	// used to setup the `http.DefaultTransport` global variable once, we need to
	// create a client here like if it was using the default created with the
	// clientTLS config.
	// This is a copy of the code we can find inside
	// InitHTTPDefaultTransportWithLogger(serviceConfig, nil):
	transport := NewTransport(cfg, logger)

	defClient := http.Client{
		Transport: transport,
	}
	resp, err = defClient.Get(fmt.Sprintf("https://localhost:%d", port))
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

func TestRunServer_MTLSOldConfigFormat(t *testing.T) {
	testKeysAreAvailable(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := 36517
	done := make(chan error)

	cfg := config.ServiceConfig{
		Port: port,
		TLS: &config.TLS{
			PublicKey:  "cert.pem",
			PrivateKey: "key.pem",
			CaCerts:    []string{"ca.pem"},
			EnableMTLS: true,
		},
		ClientTLS: &config.ClientTLS{
			AllowInsecureConnections: false, // we do not check the server cert
			CaCerts:                  []string{"ca.pem"},
			ClientCerts: []config.ClientTLSCert{
				{
					Certificate: "cert.pem",
					PrivateKey:  "key.pem",
				},
			},
		},
	}
	go func() {
		done <- RunServer(ctx, cfg, http.HandlerFunc(dummyHandler))
	}()

	client, err := mtlsClient("cert.pem", "key.pem")
	if err != nil {
		t.Error(err)
		return
	}

	<-time.After(1000 * time.Millisecond)

	resp, err := client.Get(fmt.Sprintf("https://localhost:%d", port))
	if err != nil {
		t.Error(err)
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	logger := logging.NoOp
	// since test are run in a suite, and `InitHTTPDefaultTransportWithLogger` is
	// used to setup the `http.DefaultTransport` global variable once, we need to
	// create a client here like if it was using the default created with the
	// clientTLS config.
	// This is a copy of the code we can find inside
	// InitHTTPDefaultTransportWithLogger(serviceConfig, nil):
	transport := NewTransport(cfg, logger)

	defClient := http.Client{
		Transport: transport,
	}
	resp, err = defClient.Get(fmt.Sprintf("https://localhost:%d", port))
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

func TestRunServer_h2c(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := newPort()

	done := make(chan error)
	go func() {
		done <- RunServer(
			ctx,
			config.ServiceConfig{
				Port:   port,
				UseH2C: true,
			},
			http.HandlerFunc(dummyHandler),
		)
	}()

	<-time.After(100 * time.Millisecond)

	client := h2cClient()
	resp, err := client.Get(fmt.Sprintf("http://localhost:%d", port))
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

func Test_parseTLSVersion(t *testing.T) {
	for _, tc := range []struct {
		in  string
		out uint16
	}{
		{in: "SSL3.0", out: tls.VersionSSL30},
		{in: "TLS10", out: tls.VersionTLS10},
		{in: "TLS11", out: tls.VersionTLS11},
		{in: "TLS12", out: tls.VersionTLS12},
		{in: "TLS13", out: tls.VersionTLS13},
		{in: "Unknown", out: tls.VersionTLS13},
	} {
		if res := parseTLSVersion(tc.in); res != tc.out {
			t.Errorf("input %s generated output %d. expected: %d", tc.in, res, tc.out)
		}
	}
}

func Test_parseCurveIDs(t *testing.T) {
	original := []uint16{1, 2, 3}
	cs := parseCurveIDs(original)
	for k, v := range cs {
		if original[k] != uint16(v) {
			t.Errorf("unexpected curves %v. expected: %v", cs, original)
		}
	}
}

func Test_parseCipherSuites(t *testing.T) {
	original := []uint16{1, 2, 3}
	cs := parseCipherSuites(original)
	for k, v := range cs {
		if original[k] != uint16(v) {
			t.Errorf("unexpected ciphersuites %v. expected: %v", cs, original)
		}
	}
}

func dummyHandler(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "Hello, %q", html.EscapeString(req.URL.Path))
}

func testKeysAreAvailable(t *testing.T) {
	files, err := os.ReadDir(".")
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
	cer, err := os.ReadFile(cert)
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

func mtlsClient(certPath, keyPath string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	cacer, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(cacer)
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
		RootCAs:      roots,
		Certificates: []tls.Certificate{cert},
	}
	return &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConf}}, nil
}

// h2cClient initializes client which executes cleartext http2 requests
func h2cClient() *http.Client {
	return &http.Client{
		Transport: &http2.Transport{
			DialTLSContext: func(_ context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
			AllowHTTP: true,
		},
	}
}

// newPort returns random port numbers to avoid port collisions during the tests
func newPort() int {
	return 16666 + rand.Intn(40000) // skipcq: GSC-G404
}

func TestRunServer_MultipleTLS(t *testing.T) {
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
					CaCerts: []string{"ca.pem", "exampleca.pem"},
					Keys: []config.TLSKeyPair{
						{
							PublicKey:  "cert.pem",
							PrivateKey: "key.pem",
						},
						{
							PublicKey:  "examplecert.pem",
							PrivateKey: "examplekey.pem",
						},
					},
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

	client, err = httpsClient("examplecert.pem")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = client.Get(fmt.Sprintf("https://127.0.0.1:%d", port))
	// should fail, because it will be served with cert.pem
	if err == nil || strings.Contains(err.Error(), "bad certificate") {
		t.Error("expected to have 'bad certificate' error")
		return
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://example.com:%d", port), http.NoBody)
	overrideHostTransport(client)
	resp, err = client.Do(req)
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

// overrideHostTransport subtitutes the actual address that the request will
// connecto (overriding the dns resolution).
func overrideHostTransport(client *http.Client) {
	t := http.DefaultTransport.(*http.Transport).Clone()
	if client.Transport != nil {
		if tt, ok := client.Transport.(*http.Transport); ok {
			t = tt
		}
	}
	myDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
	t.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		_, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		overrideAddress := net.JoinHostPort("127.0.0.1", port)
		return myDialer.DialContext(ctx, network, overrideAddress)
	}
	client.Transport = t
}
