package mux

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
)

func TestEndpointHandler_ok(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"supu": "tupu"},
		}, nil
	}
	expectedBody := "{\"supu\":\"tupu\"}"
	testEndpointHandler(t, 10, p, "GET", expectedBody, "public, max-age=21600", "application/json", http.StatusOK)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_incomplete(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: false,
			Data:       map[string]interface{}{"foo": "bar"},
		}, nil
	}
	expectedBody := "{\"foo\":\"bar\"}"
	testEndpointHandler(t, 10, p, "GET", expectedBody, "", "application/json", http.StatusOK)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_ko(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, fmt.Errorf("This is %s", "a dummy error")
	}
	testEndpointHandler(t, 10, p, "GET", "This is a dummy error\n", "", "text/plain; charset=utf-8", http.StatusInternalServerError)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_cancel(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	}
	testEndpointHandler(t, 0, p, "GET", ErrInternalError.Error()+"\n", "", "text/plain; charset=utf-8", http.StatusInternalServerError)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_noop(t *testing.T) {
	testEndpointHandler(t, 10, proxy.NoopProxy, "GET", "", "", "application/json", http.StatusOK)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_badMethod(t *testing.T) {
	testEndpointHandler(t, 10, proxy.NoopProxy, "PUT", "\n", "", "text/plain; charset=utf-8", http.StatusMethodNotAllowed)
	time.Sleep(5 * time.Millisecond)
}

func testEndpointHandler(t *testing.T, timeout time.Duration, p proxy.Proxy, method, expectedBody, expectedCache, expectedContent string,
	expectedStatusCode int) {
	endpoint := &config.EndpointConfig{
		Method:      "GET",
		Timeout:     timeout,
		CacheTTL:    6 * time.Hour,
		QueryString: []string{"b"},
	}

	server := startMuxServer(EndpointHandler(endpoint, p))
	defer server.Shutdown(context.Background())

	time.Sleep(15 * time.Millisecond)

	req, _ := http.NewRequest(method, "http://127.0.0.1:8081/_mux_endpoint?b=1", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Making the request:", err.Error())
		return
	}
	defer resp.Body.Close()

	body, ioerr := ioutil.ReadAll(resp.Body)
	if ioerr != nil {
		t.Error("Reading the response:", ioerr.Error())
		return
	}
	content := string(body)
	if resp.Header.Get("Cache-Control") != expectedCache {
		t.Error("Cache-Control error:", resp.Header.Get("Cache-Control"))
	}
	if resp.Header.Get("Content-Type") != expectedContent {
		t.Error("Content-Type error:", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("X-Krakend") != "Version undefined" {
		t.Error("X-Krakend error:", resp.Header.Get("X-Krakend"))
	}
	if resp.StatusCode != expectedStatusCode {
		t.Error("Unexpected status code:", resp.StatusCode)
	}
	if content != expectedBody {
		t.Error("Unexpected body:", content, "expected:", expectedBody)
	}
}

func startMuxServer(handlerFunc http.HandlerFunc) *http.Server {
	router := http.NewServeMux()
	router.Handle("/_mux_endpoint", handlerFunc)
	s := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	go s.ListenAndServe()
	return s
}
