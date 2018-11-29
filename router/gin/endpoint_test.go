package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
)

func TestEndpointHandler_ok(t *testing.T) {
	p := func(ctx context.Context, req *proxy.Request) (*proxy.Response, error) {
		if v, ok := ctx.Value("bool").(bool); !ok || !v {
			t.Errorf("unexpected bool context value: %v", v)
		}
		if v, ok := ctx.Value("int").(int); !ok || v != 42 {
			t.Errorf("unexpected int context value: %v", v)
		}
		if v, ok := ctx.Value("string").(string); !ok || v != "supu" {
			t.Errorf("unexpected string context value: %v", v)
		}
		data, _ := json.Marshal(req.Query)
		if string(data) != `{"b":["1"],"c[]":["x","y"],"d":["1","2"]}` {
			t.Errorf("unexpected querystring: %s", data)
		}
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"supu": "tupu"},
			Metadata: proxy.Metadata{
				Headers: map[string][]string{"a": {"a1", "a2"}},
			},
		}, nil
	}
	expectedBody := "{\"supu\":\"tupu\"}"
	testEndpointHandler(t, 10, p, expectedBody, "public, max-age=21600", "application/json; charset=utf-8", http.StatusOK, true)
}

var ctxContent = map[string]interface{}{
	"bool":   true,
	"int":    42,
	"string": "supu",
}

func TestEndpointHandler_incomplete(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: false,
			Data:       map[string]interface{}{"foo": "bar"},
		}, nil
	}
	expectedBody := "{\"foo\":\"bar\"}"
	testEndpointHandler(t, 10, p, expectedBody, "", "application/json; charset=utf-8", http.StatusOK, false)
}

func TestEndpointHandler_ko(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, errors.New("This is a dummy error")
	}
	testEndpointHandler(t, 10, p, "", "", "", http.StatusInternalServerError, false)
}

func TestEndpointHandler_errored(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, errors.New("this is a dummy error")
	}
	testEndpointHandler(t, 10, p, "", "", "", http.StatusInternalServerError, false)
}

func TestEndpointHandler_errored_responseError(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, dummyResponseError{err: "this is a dummy error", status: http.StatusTeapot}
	}
	testEndpointHandler(t, 10, p, "", "", "", http.StatusTeapot, false)
}

type dummyResponseError struct {
	err    string
	status int
}

func (d dummyResponseError) Error() string {
	return d.err
}

func (d dummyResponseError) StatusCode() int {
	return d.status
}

func TestEndpointHandler_incompleteAndErrored(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: false,
			Data:       map[string]interface{}{"foo": "bar"},
		}, errors.New("This is a dummy error")
	}
	expectedBody := "{\"foo\":\"bar\"}"
	testEndpointHandler(t, 10, p, expectedBody, "", "application/json; charset=utf-8", http.StatusOK, false)
}

func TestEndpointHandler_cancelEmpty(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	}
	testEndpointHandler(t, 0, p, "", "", "", http.StatusInternalServerError, false)
}

func TestEndpointHandler_cancel(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		time.Sleep(100 * time.Millisecond)
		return &proxy.Response{
			IsComplete: false,
			Data:       map[string]interface{}{"foo": "bar"},
		}, nil
	}
	expectedBody := "{\"foo\":\"bar\"}"
	testEndpointHandler(t, 0, p, expectedBody, "", "application/json; charset=utf-8", http.StatusOK, false)
}

func TestEndpointHandler_noop(t *testing.T) {
	testEndpointHandler(t, time.Minute, proxy.NoopProxy, "{}", "", "application/json; charset=utf-8", http.StatusOK, false)
}

func testEndpointHandler(t *testing.T, timeout time.Duration, p proxy.Proxy, expectedBody, expectedCache,
	expectedContent string, expectedStatusCode int, completed bool) {
	body, resp, err := setup(timeout, p)
	if err != nil {
		t.Error("Reading the response:", err.Error())
		return
	}
	content := string(body)
	if resp.Header.Get("Cache-Control") != expectedCache {
		t.Error("Cache-Control error:", resp.Header.Get("Cache-Control"))
	}
	if completed && resp.Header.Get(router.CompleteResponseHeaderName) != router.HeaderCompleteResponseValue {
		t.Error(router.CompleteResponseHeaderName, "error:", resp.Header.Get(router.CompleteResponseHeaderName))
	}
	if !completed && resp.Header.Get(router.CompleteResponseHeaderName) != router.HeaderIncompleteResponseValue {
		t.Error(router.CompleteResponseHeaderName, "error:", resp.Header.Get(router.CompleteResponseHeaderName))
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
		QueryString: []string{"b", "c[]", "d"},
	}

	server := startGinServer(EndpointHandler(endpoint, p))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1&c[]=x&c[]=y&d=1&d=2", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	server.ServeHTTP(w, req)

	defer w.Result().Body.Close()

	body, ioerr := ioutil.ReadAll(w.Result().Body)
	if ioerr != nil {
		return "", nil, ioerr
	}
	return string(body), w.Result(), nil
}

func startGinServer(handlerFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/_gin_endpoint/:param", ctxMiddleware, handlerFunc)

	return router
}

func ctxMiddleware(c *gin.Context) {
	for k, v := range ctxContent {
		c.Set(k, v)
	}
}
