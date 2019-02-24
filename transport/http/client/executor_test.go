//go:generate go run $GOROOT/src/crypto/tls/generate_cert.go --rsa-bits 1024 --host 127.0.0.1,::1,localhost --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h
package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestDefaultHTTPRequestExecutor(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	re := DefaultHTTPRequestExecutor(NewHTTPClient)

	req, _ := http.NewRequest("GET", ts.URL, ioutil.NopCloser(&bytes.Buffer{}))

	resp, err := re(context.Background(), req)

	if err != nil {
		t.Error("unexpected error:", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", resp.StatusCode)
	}
}

func TestInitHTTPDefaultTransport(t *testing.T) {
	testKeysAreAvailable(t)

	cleanUp()
	defer cleanUp()

	err := InitHTTPDefaultTransport(config.ServiceConfig{TLS: &config.TLS{
		LocalCA: "cert.pem",
	}})
	if err != nil {
		t.Error(err)
		return
	}

	port := newPort()
	close := createAndStartHTTPSServer(port)
	defer close()

	<-time.After(200 * time.Millisecond)
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/", port), nil)
	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		t.Error(err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status code. have: %d, want: 200", resp.StatusCode)
	}
}

func TestInitHTTPDefaultTransport_unknownCA(t *testing.T) {
	testKeysAreAvailable(t)

	cleanUp()
	defer cleanUp()

	err := InitHTTPDefaultTransport(config.ServiceConfig{})
	if err != nil {
		t.Error(err)
		return
	}

	port := newPort()
	close := createAndStartHTTPSServer(port)
	defer close()

	<-time.After(200 * time.Millisecond)
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://localhost:%d/", port), nil)
	resp, err := defaultHTTPClient.Do(req)
	expectedErr := fmt.Sprintf("Get https://localhost:%d/: x509: certificate signed by unknown authority", port)
	if err == nil || err.Error() != expectedErr {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if resp != nil {
		t.Errorf("unexpected response. have: %v", resp)
	}
}

func Test_newHTTPClient_unknownCAPath(t *testing.T) {
	testKeysAreAvailable(t)

	client, err := newHTTPClient(config.ServiceConfig{TLS: &config.TLS{LocalCA: "no_such_file"}})
	if err == nil || err.Error() != "Failed to append no_such_file to RootCAs: open no_such_file: no such file or directory" {
		t.Errorf("unexpected error: %v", err)
	}

	if client != nil {
		t.Errorf("unexpected client %v", client)
	}
}

func testKeysAreAvailable(t *testing.T) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		t.Fatal(err)
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

func newPort() int {
	return 16600 + rand.Intn(40000)
}

func createAndStartHTTPSServer(port int) func() error {
	s := http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte("ok"))
		}),
	}

	go func() {
		s.ListenAndServeTLS("cert.pem", "key.pem")
	}()

	return s.Close
}

func cleanUp() {
	once = sync.Once{}
	defaultHTTPClient = &http.Client{}
}
