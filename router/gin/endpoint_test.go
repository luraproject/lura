package gin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

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
	testEndpointHandler(t, 10, p, expectedBody, "public, max-age=21600", "application/json; charset=utf-8", http.StatusOK)
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
	testEndpointHandler(t, 10, p, expectedBody, "", "application/json; charset=utf-8", http.StatusOK)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_ko(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, fmt.Errorf("This is %s", "a dummy error")
	}
	testEndpointHandler(t, 10, p, "", "", "text/plain; charset=utf-8", http.StatusInternalServerError)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_cancel(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	}
	testEndpointHandler(t, 0, p, "{}", "", "text/plain; charset=utf-8", http.StatusInternalServerError)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_noop(t *testing.T) {
	testEndpointHandler(t, 10, proxy.NoopProxy, "{}", "", "application/json; charset=utf-8", http.StatusOK)
	time.Sleep(5 * time.Millisecond)
}

func testEndpointHandler(t *testing.T, timeout time.Duration, p proxy.Proxy, expectedBody, expectedCache,
	expectedContent string, expectedStatusCode int) {
	body, resp, err := setup(timeout, p)
	if err != nil {
		t.Error("Reading the response:", err.Error())
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

func setup(timeout time.Duration, p proxy.Proxy) (string, *http.Response, error) {
	endpoint := &config.EndpointConfig{
		Timeout:     timeout,
		CacheTTL:    6 * time.Hour,
		QueryString: []string{"b"},
	}

	server := startGinServer(EndpointHandler(endpoint, p))
	defer server.Shutdown(context.Background())

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	body, ioerr := ioutil.ReadAll(resp.Body)
	if ioerr != nil {
		return "", nil, err
	}
	return string(body), resp, nil
}

func startGinServer(handlerFunc gin.HandlerFunc) *http.Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/_gin_endpoint/:param", handlerFunc)
	s := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	go s.ListenAndServe()
	time.Sleep(5 * time.Millisecond)
	return s
}
