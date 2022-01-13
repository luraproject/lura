// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/encoding"
	"github.com/luraproject/lura/v2/proxy"
)

func TestRender_Negotiated_ok(t *testing.T) {
	type A struct {
		B string
	}
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"content": A{B: "supu"}},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: NEGOTIATE,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain", "application/x-yaml; charset=utf-8", "content:\n  b: supu\n"},
		{"none", "", "application/json; charset=utf-8", `{"content":{"B":"supu"}}`},
		{"json", "application/json", "application/json; charset=utf-8", `{"content":{"B":"supu"}}`},
		{"xml", "application/xml", "application/xml; charset=utf-8", `<A><B>supu</B></A>`},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer w.Result().Body.Close()

		body, ioerr := ioutil.ReadAll(w.Result().Body)
		if ioerr != nil {
			t.Error("reading response body:", ioerr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Cache-Control") != "public, max-age=21600" {
			t.Error(testData[0], "Cache-Control error:", w.Result().Header.Get("Cache-Control"))
		}
		if w.Result().Header.Get("Content-Type") != testData[2] {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Krakend") != "Version undefined" {
			t.Error(testData[0], "X-Krakend error:", w.Result().Header.Get("X-Krakend"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != testData[3] {
			t.Error(testData[0], "Unexpected body:", content, "expected:", testData[3])
		}
	}
}

func TestRender_Negotiated_noData(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			Data: map[string]interface{}{},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: NEGOTIATE,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain", "application/x-yaml; charset=utf-8", "{}\n"},
		{"none", "", "application/json; charset=utf-8", "{}"},
		{"json", "application/json", "application/json; charset=utf-8", "{}"},
		{"xml", "application/xml", "application/xml; charset=utf-8", ""},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer w.Result().Body.Close()

		body, ioerr := ioutil.ReadAll(w.Result().Body)
		if ioerr != nil {
			t.Error("reading response body:", ioerr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Content-Type") != testData[2] {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Krakend") != "Version undefined" {
			t.Error(testData[0], "X-Krakend error:", w.Result().Header.Get("X-Krakend"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != testData[3] {
			t.Error(testData[0], "Unexpected body:", content, "expected:", testData[3])
		}
	}
}

func TestRender_Negotiated_noResponse(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: NEGOTIATE,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain", "application/x-yaml; charset=utf-8", "{}\n"},
		{"none", "", "application/json; charset=utf-8", "{}"},
		{"json", "application/json", "application/json; charset=utf-8", "{}"},
		{"xml", "application/xml", "application/xml; charset=utf-8", ""},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer w.Result().Body.Close()

		body, ioerr := ioutil.ReadAll(w.Result().Body)
		if ioerr != nil {
			t.Error("reading response body:", ioerr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Content-Type") != testData[2] {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Krakend") != "Version undefined" {
			t.Error(testData[0], "X-Krakend error:", w.Result().Header.Get("X-Krakend"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != testData[3] {
			t.Error(testData[0], "Unexpected body:", content, "expected:", testData[3])
		}
	}
}

func TestRender_unknown(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"supu": "tupu"},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: "unknown",
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	expectedHeader := "application/json; charset=utf-8"
	expectedBody := `{"supu":"tupu"}`

	for _, testData := range [][]string{
		{"plain", "text/plain"},
		{"none", ""},
		{"json", "application/json"},
		{"unknown", "unknown"},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer w.Result().Body.Close()

		body, ioerr := ioutil.ReadAll(w.Result().Body)
		if ioerr != nil {
			t.Error("reading response body:", ioerr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Cache-Control") != "public, max-age=21600" {
			t.Error(testData[0], "Cache-Control error:", w.Result().Header.Get("Cache-Control"))
		}
		if w.Result().Header.Get("Content-Type") != expectedHeader {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Krakend") != "Version undefined" {
			t.Error(testData[0], "X-Krakend error:", w.Result().Header.Get("X-Krakend"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != expectedBody {
			t.Error(testData[0], "Unexpected body:", content, "expected:", expectedBody)
		}
	}
}

func TestRender_string(t *testing.T) {
	expectedContent := "supu"
	expectedHeader := "text/plain; charset=utf-8"

	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       map[string]interface{}{"content": expectedContent},
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.STRING,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	for _, testData := range [][]string{
		{"plain", "text/plain"},
		{"none", ""},
		{"json", "application/json"},
		{"unknown", "unknown"},
	} {
		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", testData[1])

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer w.Result().Body.Close()

		body, ioerr := ioutil.ReadAll(w.Result().Body)
		if ioerr != nil {
			t.Error("reading response body:", ioerr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Cache-Control") != "public, max-age=21600" {
			t.Error(testData[0], "Cache-Control error:", w.Result().Header.Get("Cache-Control"))
		}
		if w.Result().Header.Get("Content-Type") != expectedHeader {
			t.Error(testData[0], "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Krakend") != "Version undefined" {
			t.Error(testData[0], "X-Krakend error:", w.Result().Header.Get("X-Krakend"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(testData[0], "Unexpected status code:", w.Result().StatusCode)
		}
		if content != expectedContent {
			t.Error(testData[0], "Unexpected body:", content, "expected:", expectedContent)
		}
	}
}

func TestRender_string_noData(t *testing.T) {
	expectedContent := ""
	expectedHeader := "text/plain; charset=utf-8"

	for k, p := range []proxy.Proxy{
		func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{
				IsComplete: false,
				Data:       map[string]interface{}{"content": 42},
			}, nil
		},
		func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return &proxy.Response{
				IsComplete: false,
				Data:       map[string]interface{}{},
			}, nil
		},
		func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
			return nil, nil
		},
	} {
		endpoint := &config.EndpointConfig{
			Timeout:        time.Second,
			CacheTTL:       6 * time.Hour,
			QueryString:    []string{"b"},
			OutputEncoding: encoding.STRING,
		}

		gin.SetMode(gin.TestMode)
		server := gin.New()
		server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

		req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		defer w.Result().Body.Close()

		body, ioerr := ioutil.ReadAll(w.Result().Body)
		if ioerr != nil {
			t.Error("reading response body:", ioerr)
			return
		}

		content := string(body)
		if w.Result().Header.Get("Content-Type") != expectedHeader {
			t.Error(k, "Content-Type error:", w.Result().Header.Get("Content-Type"))
		}
		if w.Result().Header.Get("X-Krakend") != "Version undefined" {
			t.Error(k, "X-Krakend error:", w.Result().Header.Get("X-Krakend"))
		}
		if w.Result().StatusCode != http.StatusOK {
			t.Error(k, "Unexpected status code:", w.Result().StatusCode)
		}
		if content != expectedContent {
			t.Error(k, "Unexpected body:", content, "expected:", expectedContent)
		}
	}
}

func TestRegisterRender(t *testing.T) {
	var total int
	expected := &proxy.Response{IsComplete: true, Data: map[string]interface{}{"a": "b"}}
	name := "test render"

	RegisterRender(name, func(_ *gin.Context, resp *proxy.Response) {
		*resp = *expected
		total++
	})

	subject := getRender(&config.EndpointConfig{OutputEncoding: name})

	var c *gin.Context
	resp := proxy.Response{}
	subject(c, &resp)

	if !reflect.DeepEqual(resp, *expected) {
		t.Error("unexpected response", resp)
	}

	if total != 1 {
		t.Error("the render was called an unexpected amount of times:", total)
	}
}

func TestRender_noop(t *testing.T) {
	expectedContent := "supu"
	expectedHeader := "text/plain; charset=utf-8"
	expectedSetCookieValue := []string{"test1=test1", "test2=test2"}

	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			Metadata: proxy.Metadata{
				StatusCode: 200,
				Headers: map[string][]string{
					"Content-Type": {expectedHeader},
					"Set-Cookie":   {"test1=test1", "test2=test2"},
				},
			},
			Io: bytes.NewBufferString(expectedContent),
		}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.NOOP,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	defer w.Result().Body.Close()

	body, ioerr := ioutil.ReadAll(w.Result().Body)
	if ioerr != nil {
		t.Error("reading response body:", ioerr)
		return
	}

	content := string(body)
	if w.Result().Header.Get("Content-Type") != expectedHeader {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Krakend") != "Version undefined" {
		t.Error("X-Krakend error:", w.Result().Header.Get("X-Krakend"))
	}
	if w.Result().StatusCode != http.StatusOK {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
	if content != expectedContent {
		t.Error("Unexpected body:", content, "expected:", expectedContent)
	}
	gotCookie := w.Header()["Set-Cookie"]
	if !reflect.DeepEqual(gotCookie, expectedSetCookieValue) {
		t.Error("Unexpected Set-Cookie header:", gotCookie, "expected:", expectedSetCookieValue)
	}
}

func TestRender_noop_nilBody(t *testing.T) {
	expectedContent := ""
	expectedHeader := ""

	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{IsComplete: true}, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.NOOP,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	defer w.Result().Body.Close()

	body, ioerr := ioutil.ReadAll(w.Result().Body)
	if ioerr != nil {
		t.Error("reading response body:", ioerr)
		return
	}

	content := string(body)
	if w.Result().Header.Get("Content-Type") != expectedHeader {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Krakend") != "Version undefined" {
		t.Error("X-Krakend error:", w.Result().Header.Get("X-Krakend"))
	}
	if w.Result().StatusCode != http.StatusOK {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
	if content != expectedContent {
		t.Error("Unexpected body:", content, "expected:", expectedContent)
	}
}

func TestRender_noop_nilResponse(t *testing.T) {
	p := func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return nil, nil
	}
	endpoint := &config.EndpointConfig{
		Timeout:        time.Second,
		CacheTTL:       6 * time.Hour,
		QueryString:    []string{"b"},
		OutputEncoding: encoding.NOOP,
	}

	gin.SetMode(gin.TestMode)
	server := gin.New()
	server.GET("/_gin_endpoint/:param", EndpointHandler(endpoint, p))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)

	if w.Result().Header.Get("Content-Type") != "" {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Krakend") != "Version undefined" {
		t.Error("X-Krakend error:", w.Result().Header.Get("X-Krakend"))
	}
	if w.Result().StatusCode != http.StatusInternalServerError {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
}
