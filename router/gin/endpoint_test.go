// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/transport/http/server"
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

	endpointHandlerTestCase{
		timeout:            10,
		proxy:              p,
		method:             "GET",
		expectedBody:       "{\"supu\":\"tupu\"}",
		expectedCache:      "public, max-age=21600",
		expectedContent:    "application/json; charset=utf-8",
		expectedStatusCode: http.StatusOK,
		completed:          true,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_okAllParams(t *testing.T) {
	p := func(_ context.Context, req *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"query": req.Query, "headers": req.Headers, "params": req.Params},
			Metadata: proxy.Metadata{
				Headers:    map[string][]string{"X-YZ": {"something"}},
				StatusCode: 200,
			},
		}, nil
	}
	endpointHandlerTestCase{
		timeout:            10,
		proxy:              p,
		method:             "GET",
		expectedBody:       `{"headers":{"Content-Type":["application/json"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-For":[""],"X-Forwarded-Host":["127.0.0.1:8080"]},"params":{"Param":"a"},"query":{"a":["42"],"b":["1"],"c[]":["x","y"],"d":["1","2"]}}`,
		expectedCache:      "public, max-age=21600",
		expectedContent:    "application/json; charset=utf-8",
		expectedStatusCode: http.StatusOK,
		completed:          true,
		queryString:        []string{"*"},
		headers:            []string{"*"},
		expectedHeaders:    map[string][]string{"X-YZ": {"something"}},
	}.test(t)
	time.Sleep(5 * time.Millisecond)
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
	endpointHandlerTestCase{
		timeout:            10,
		proxy:              p,
		method:             "GET",
		expectedBody:       "{\"foo\":\"bar\"}",
		expectedCache:      "",
		expectedContent:    "application/json; charset=utf-8",
		expectedStatusCode: http.StatusOK,
		completed:          false,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_errored(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, errors.New("this is a dummy error")
	}
	endpointHandlerTestCase{
		timeout:            10,
		proxy:              p,
		method:             "GET",
		expectedBody:       "",
		expectedCache:      "",
		expectedContent:    "",
		expectedStatusCode: http.StatusInternalServerError,
		completed:          false,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_errored_responseError(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, dummyResponseError{err: "this is a dummy error", status: http.StatusTeapot}
	}
	endpointHandlerTestCase{
		timeout:            10,
		proxy:              p,
		method:             "GET",
		expectedBody:       "",
		expectedCache:      "",
		expectedContent:    "",
		expectedStatusCode: http.StatusTeapot,
		completed:          false,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
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
	endpointHandlerTestCase{
		timeout:            10,
		proxy:              p,
		method:             "GET",
		expectedBody:       "{\"foo\":\"bar\"}",
		expectedCache:      "",
		expectedContent:    "application/json; charset=utf-8",
		expectedStatusCode: http.StatusOK,
		completed:          false,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_cancelEmpty(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	}
	endpointHandlerTestCase{
		timeout:            0,
		proxy:              p,
		method:             "GET",
		expectedBody:       "",
		expectedCache:      "",
		expectedContent:    "",
		expectedStatusCode: http.StatusInternalServerError,
		completed:          false,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_cancel(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		time.Sleep(100 * time.Millisecond)
		return &proxy.Response{
			IsComplete: false,
			Data:       map[string]interface{}{"foo": "bar"},
		}, nil
	}
	endpointHandlerTestCase{
		timeout:            0,
		proxy:              p,
		method:             "GET",
		expectedBody:       "{\"foo\":\"bar\"}",
		expectedCache:      "",
		expectedContent:    "application/json; charset=utf-8",
		expectedStatusCode: http.StatusOK,
		completed:          false,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
}

func TestEndpointHandler_noop(t *testing.T) {
	endpointHandlerTestCase{
		timeout:            time.Minute,
		proxy:              proxy.NoopProxy,
		method:             "GET",
		expectedBody:       "{}",
		expectedCache:      "",
		expectedContent:    "application/json; charset=utf-8",
		expectedStatusCode: http.StatusOK,
		completed:          false,
	}.test(t)
	time.Sleep(5 * time.Millisecond)
}

func TestCustomErrorEndpointHandler(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := logging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}
	hf := CustomErrorEndpointHandler(logger, server.DefaultToHTTPError)

	endpoint := &config.EndpointConfig{
		Method:      "GET",
		Endpoint:    "/",
		Timeout:     time.Minute,
		CacheTTL:    6 * time.Hour,
		QueryString: []string{"b", "c[]", "d"},
	}

	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, errors.New("this is a dummy error")
	}

	s := startGinServer(hf(endpoint, p))

	req, _ := http.NewRequest(
		"GET",
		"http://127.0.0.1:8080/_gin_endpoint/a?a=42&b=1&c[]=x&c[]=y&d=1&d=2",
		ioutil.NopCloser(&bytes.Buffer{}),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	if content := buff.String(); !strings.Contains(content, "pref ERROR: [ENDPOINT: /] this is a dummy error") {
		t.Error("unexpected log content", content)
	}
}

type endpointHandlerTestCase struct {
	timeout            time.Duration
	proxy              proxy.Proxy
	method             string
	expectedBody       string
	expectedCache      string
	expectedContent    string
	expectedHeaders    map[string][]string
	expectedStatusCode int
	completed          bool
	queryString        []string
	headers            []string
}

func (tc endpointHandlerTestCase) test(t *testing.T) {
	endpoint := &config.EndpointConfig{
		Method:      "GET",
		Timeout:     tc.timeout,
		CacheTTL:    6 * time.Hour,
		QueryString: []string{"b", "c[]", "d"},
	}
	if len(tc.queryString) > 0 {
		endpoint.QueryString = tc.queryString
	}
	if len(tc.headers) > 0 {
		endpoint.HeadersToPass = tc.headers
	}

	s := startGinServer(EndpointHandler(endpoint, tc.proxy))

	req, _ := http.NewRequest(
		tc.method,
		"http://127.0.0.1:8080/_gin_endpoint/a?a=42&b=1&c[]=x&c[]=y&d=1&d=2",
		ioutil.NopCloser(&bytes.Buffer{}),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	s.ServeHTTP(w, req)

	body, ioerr := ioutil.ReadAll(w.Result().Body)
	if ioerr != nil {
		t.Error("Reading the response:", ioerr.Error())
		return
	}
	w.Result().Body.Close()
	content := string(body)
	resp := w.Result()
	if resp.Header.Get("Cache-Control") != tc.expectedCache {
		t.Error("Cache-Control error:", resp.Header.Get("Cache-Control"))
	}
	if tc.completed && resp.Header.Get(server.CompleteResponseHeaderName) != server.HeaderCompleteResponseValue {
		t.Error(server.CompleteResponseHeaderName, "error:", resp.Header.Get(server.CompleteResponseHeaderName))
	}
	if !tc.completed && resp.Header.Get(server.CompleteResponseHeaderName) != server.HeaderIncompleteResponseValue {
		t.Error(server.CompleteResponseHeaderName, "error:", resp.Header.Get(server.CompleteResponseHeaderName))
	}
	if resp.Header.Get("Content-Type") != tc.expectedContent {
		t.Error("Content-Type error:", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("X-Krakend") != "Version undefined" {
		t.Error("X-Krakend error:", resp.Header.Get("X-Krakend"))
	}
	if resp.StatusCode != tc.expectedStatusCode {
		t.Error("Unexpected status code:", resp.StatusCode)
	}
	if content != tc.expectedBody {
		t.Error("Unexpected body:", content, "expected:", tc.expectedBody)
	}
	for k, v := range tc.expectedHeaders {
		if header := resp.Header.Get(k); v[0] != header {
			t.Error("Unexpected value for header:", k, header, "expected:", v[0])
		}
	}
}

func startGinServer(handlerFunc gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/_gin_endpoint/:param", ctxMiddleware, handlerFunc)

	return r
}

func ctxMiddleware(c *gin.Context) {
	for k, v := range ctxContent {
		c.Set(k, v)
	}
}
