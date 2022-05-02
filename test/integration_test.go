//go:build integration || !race
// +build integration !race

// SPDX-License-Identifier: Apache-2.0

package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
	"time"

	ginlib "github.com/gin-gonic/gin"
	"github.com/urfave/negroni/v2"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/router/chi"
	"github.com/luraproject/lura/v2/router/gin"
	"github.com/luraproject/lura/v2/router/gorilla"
	"github.com/luraproject/lura/v2/router/httptreemux"
	luranegroni "github.com/luraproject/lura/v2/router/negroni"
	"github.com/luraproject/lura/v2/transport/http/server"
)

func TestKrakenD_ginRouter(t *testing.T) {
	ginlib.SetMode(ginlib.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testKrakenD(t, func(logger logging.Logger, cfg *config.ServiceConfig) {
		if cfg.ExtraConfig == nil {
			cfg.ExtraConfig = map[string]interface{}{}
		}
		cfg.ExtraConfig[gin.Namespace] = map[string]interface{}{
			"trusted_proxies":        []interface{}{"127.0.0.1/32", "::1"},
			"remote_ip_headers":      []interface{}{"x-forwarded-for"},
			"forwarded_by_client_ip": true,
		}

		ignoredChan := make(chan string)
		opts := gin.EngineOptions{
			Logger: logger,
			Writer: ioutil.Discard,
			Health: (<-chan string)(ignoredChan),
		}

		gin.NewFactory(
			gin.Config{
				Engine:         gin.NewEngine(*cfg, opts),
				Middlewares:    []ginlib.HandlerFunc{},
				HandlerFactory: gin.EndpointHandler,
				ProxyFactory:   proxy.DefaultFactory(logger),
				Logger:         logger,
				RunServer:      server.RunServer,
			},
		).NewWithContext(ctx).Run(*cfg)
	})
}

func TestKrakenD_gorillaRouter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.RoutingPattern = config.BracketsRouterPatternBuilder
	testKrakenD(t, func(logger logging.Logger, cfg *config.ServiceConfig) {
		gorilla.DefaultFactory(proxy.DefaultFactory(logger), logger).NewWithContext(ctx).Run(*cfg)
	})
	config.RoutingPattern = config.ColonRouterPatternBuilder
}

func TestKrakenD_negroniRouter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.RoutingPattern = config.BracketsRouterPatternBuilder
	testKrakenD(t, func(logger logging.Logger, cfg *config.ServiceConfig) {
		factory := luranegroni.DefaultFactory(proxy.DefaultFactory(logger), logger, []negroni.Handler{})
		factory.NewWithContext(ctx).Run(*cfg)
	})
	config.RoutingPattern = config.ColonRouterPatternBuilder
}

func TestKrakenD_httptreemuxRouter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testKrakenD(t, func(logger logging.Logger, cfg *config.ServiceConfig) {
		httptreemux.DefaultFactory(proxy.DefaultFactory(logger), logger).NewWithContext(ctx).Run(*cfg)
	})
}

func TestKrakenD_chiRouter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config.RoutingPattern = config.BracketsRouterPatternBuilder
	testKrakenD(t, func(logger logging.Logger, cfg *config.ServiceConfig) {
		chi.DefaultFactory(proxy.DefaultFactory(logger), logger).NewWithContext(ctx).Run(*cfg)
	})
	config.RoutingPattern = config.ColonRouterPatternBuilder
}

func testKrakenD(t *testing.T, runRouter func(logging.Logger, *config.ServiceConfig)) {
	cfg, err := setupBackend(t)
	if err != nil {
		t.Error(err)
		return
	}

	logger := logging.NoOp
	go runRouter(logger, cfg)

	<-time.After(300 * time.Millisecond)

	defaultHeaders := map[string]string{
		"Content-Type":        "application/json",
		"X-KrakenD-Completed": "true",
		"X-Krakend":           "Version undefined",
	}

	incompleteHeader := map[string]string{
		"Content-Type":        "application/json",
		"X-KrakenD-Completed": "false",
		"X-Krakend":           "Version undefined",
	}

	for _, tc := range []struct {
		name          string
		url           string
		method        string
		headers       map[string]string
		body          string
		expBody       string
		expHeaders    map[string]string
		expStatusCode int
	}{
		{
			name:       "static",
			url:        "/static",
			headers:    map[string]string{},
			expHeaders: incompleteHeader,
			expBody:    `{"bar":"foobar","foo":42}`,
		},
		{
			name:   "param_forwarding",
			url:    "/param_forwarding/foo/constant/bar",
			method: "POST",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "bearer AuthorizationToken",
				"X-Y-Z":         "x-y-z",
			},
			body:       `{"foo":"bar"}`,
			expHeaders: defaultHeaders,
			expBody:    `{"path":"/foo/bar"}`,
		},
		{
			name:   "param_forwarding_2",
			url:    "/param_forwarding/foo/constant/foobar",
			method: "POST",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "bearer AuthorizationToken",
				"X-Y-Z":         "x-y-z",
			},
			body:       `{"foo":"bar"}`,
			expHeaders: defaultHeaders,
			expBody:    `{"path":"/foo/foobar"}`,
		},
		{
			name:       "timeout",
			url:        "/timeout",
			headers:    map[string]string{},
			expHeaders: incompleteHeader,
			expBody:    `{"email":"some@email.com","name":"a"}`,
		},
		{
			name:       "partial_with_static",
			url:        "/partial/static",
			headers:    map[string]string{},
			expHeaders: incompleteHeader,
			expBody:    `{"bar":"foobar","email":"some@email.com","foo":42,"name":"a"}`,
		},
		{
			name:       "partial",
			url:        "/partial",
			headers:    map[string]string{},
			expHeaders: incompleteHeader,
			expBody:    `{"email":"some@email.com","name":"a"}`,
		},
		{
			name:       "combination",
			url:        "/combination",
			headers:    map[string]string{},
			expHeaders: defaultHeaders,
			expBody:    `{"name":"a","personal_email":"some@email.com","posts":[{"body":"some content","date":"123456789"},{"body":"some other content","date":"123496789"}]}`,
		},
		{
			name:       "detail_error",
			url:        "/detail_error",
			headers:    map[string]string{},
			expHeaders: incompleteHeader,
			expBody:    `{"email":"some@email.com","error_backend_a":{"http_status_code":429,"http_body":"sad panda\n"},"name":"a"}`,
		},
		{
			name:       "querystring-params-no-params",
			url:        "/querystring-params-test/no-params?a=1&b=2&c=3",
			headers:    map[string]string{},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-Host":["localhost:%d"]},"path":"/no-params","query":{}}`, cfg.Port),
		},
		{
			name:       "querystring-params-optional-query-params",
			url:        "/querystring-params-test/query-params?a=1&b=2&c=3",
			headers:    map[string]string{},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-Host":["localhost:%d"]},"path":"/query-params","query":{"a":["1"],"b":["2"]}}`, cfg.Port),
		},
		{
			name:       "querystring-params-mandatory-query-params",
			url:        "/querystring-params-test/url-params/some?a=1&b=2&c=3",
			headers:    map[string]string{},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-Host":["localhost:%d"]},"path":"/url-params","query":{"p":["some"]}}`, cfg.Port),
		},
		{
			name:       "querystring-params-all",
			url:        "/querystring-params-test/all-params?a=1&b=2&c=3",
			headers:    map[string]string{},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-Host":["localhost:%d"]},"path":"/all-params","query":{"a":["1"],"b":["2"],"c":["3"]}}`, cfg.Port),
		},
		{
			name: "header-params-none",
			url:  "/header-params-test/no-params",
			headers: map[string]string{
				"x-Test-1": "some",
				"X-TEST-2": "none",
			},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-Host":["localhost:%d"]},"path":"/no-params","query":{}}`, cfg.Port),
		},
		{
			name: "header-params-filter",
			url:  "/header-params-test/filter-params",
			headers: map[string]string{
				"x-tESt-1": "some",
				"X-TEST-2": "none",
			},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-Host":["localhost:%d"],"X-Test-1":["some"]},"path":"/filter-params","query":{}}`, cfg.Port),
		},
		{
			name: "header-params-all",
			url:  "/header-params-test/all-params",
			headers: map[string]string{
				"x-Test-1":   "some",
				"X-TEST-2":   "none",
				"User-Agent": "KrakenD Test",
			},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Test"],"X-Forwarded-Host":["localhost:%d"],"X-Forwarded-Via":["KrakenD Version undefined"],"X-Test-1":["some"],"X-Test-2":["none"]},"path":"/all-params","query":{}}`, cfg.Port),
		},
		{
			name:       "sequential ok",
			url:        "/sequential/ok/foo",
			expHeaders: defaultHeaders,
			expBody:    `{"first":{"path":"/provider/foo","random":42},"second":{"path":"/recipient/42","random":42}}`,
		},
		{
			name: "sequential ko first",
			url:  "/sequential/ko/first/foo",
			expHeaders: map[string]string{
				"X-KrakenD-Completed": "false",
				"X-Krakend":           "Version undefined",
			},
			expStatusCode: 500,
		},
		{
			name:       "sequential ko last",
			url:        "/sequential/ko/last/foo",
			expHeaders: incompleteHeader,
			expBody:    `{"random":42}`,
		},
		{
			name:       "redirect",
			url:        "/redirect",
			expHeaders: defaultHeaders,
			expBody:    `{"path":"/","random":42}`,
		},
		{
			name:       "found",
			url:        "/found",
			expHeaders: defaultHeaders,
			expBody:    `{"path":"/","random":42}`,
		},
		{
			name:       "flatmap del",
			url:        "/flatmap/delete",
			expHeaders: defaultHeaders,
			expBody:    `{"collection":[{"body":"some content"},{"body":"some other content"}]}`,
		},
		{
			name:       "flatmap rename",
			url:        "/flatmap/rename",
			expHeaders: defaultHeaders,
			expBody:    `{"collection":[{"body":"some content","created_at":"123456789"},{"body":"some other content","created_at":"123496789"}]}`,
		},
		{
			name: "x-forwarded-for",
			url:  "/x-forwarded-for",
			headers: map[string]string{
				"x-forwarded-for": "123.45.67.89",
			},
			expHeaders: defaultHeaders,
			expBody:    fmt.Sprintf(`{"headers":{"Accept-Encoding":["gzip"],"User-Agent":["KrakenD Version undefined"],"X-Forwarded-For":["123.45.67.89"],"X-Forwarded-Host":["localhost:%d"]}}`, cfg.Port),
		},
		{
			method:     "PUT",
			name:       "sequence-accept",
			url:        "/sequence-accept",
			expHeaders: defaultHeaders,
		},
		{
			method:        "GET",
			name:          "error-status-code-1",
			url:           "/error-status-code/1",
			expStatusCode: 200,
		},
		{
			method:        "GET",
			name:          "error-status-code-2",
			url:           "/error-status-code/2",
			expStatusCode: 429,
		},
		{
			method:        "GET",
			name:          "error-status-code-3",
			url:           "/error-status-code/3",
			expStatusCode: 200,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.method == "" {
				tc.method = "GET"
			}

			var body io.Reader
			if tc.body != "" {
				body = bytes.NewBufferString(tc.body)
			}

			url := fmt.Sprintf("http://localhost:%d%s", cfg.Port, tc.url)

			r, _ := http.NewRequest(tc.method, url, body)
			for k, v := range tc.headers {
				r.Header.Add(k, v)
			}

			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				t.Error(err)
				return
			}
			if resp == nil {
				t.Errorf("%s: nil response", resp.Request.URL.Path)
				return
			}

			expectedStatusCode := http.StatusOK
			if tc.expStatusCode != 0 {
				expectedStatusCode = tc.expStatusCode
			}
			if resp.StatusCode != expectedStatusCode {
				t.Errorf("%s: unexpected status code. have: %d, want: %d", resp.Request.URL.Path, resp.StatusCode, expectedStatusCode)
			}

			for k, v := range tc.expHeaders {
				if c := resp.Header.Get(k); !strings.Contains(c, v) {
					t.Errorf("%s: unexpected header %s: %s", resp.Request.URL.Path, k, c)
				}
			}
			if tc.expBody == "" {
				return
			}

			b, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if tc.expBody != string(b) {
				t.Errorf(
					"%s: unexpected body: %s\n\t%s was expecting: %s",
					resp.Request.URL.Path,
					string(b),
					resp.Request.URL.Path,
					tc.expBody,
				)
			}
		})
	}

}

func setupBackend(t *testing.T) (*config.ServiceConfig, error) {
	data := map[string]interface{}{"port": rand.Intn(2000) + 8080}

	// param forwarding validation backend
	b1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if c := r.Header.Get("Content-Type"); c != "application/json" {
			t.Errorf("unexpected header content-type: %s", c)
			http.Error(rw, "bad content-type", 400)
			return
		}
		if c := r.Header.Get("Authorization"); c != "bearer AuthorizationToken" {
			t.Errorf("unexpected header Authorization: %s", c)
			http.Error(rw, "bad Authorization", 400)
			return
		}
		if c := r.Header.Get("X-Y-Z"); c != "x-y-z" {
			t.Errorf("unexpected header X-Y-Z: %s", c)
			http.Error(rw, "bad X-Y-Z", 400)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		if string(body) != `{"foo":"bar"}` {
			t.Errorf("unexpected request body: %s", string(body))
			return
		}
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{"path": r.URL.Path})
	}))
	data["b1"] = b1.URL

	// collection generator backend
	b2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode([]interface{}{
			map[string]interface{}{"body": "some content", "date": "123456789"},
			map[string]interface{}{"body": "some other content", "date": "123496789"},
		})
	}))
	data["b2"] = b2.URL

	// regular struct generator backend
	b3 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{"email": "some@email.com", "name": "a"})
	}))
	data["b3"] = b3.URL

	// crasher backend
	b4 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Error(rw, "sad panda", http.StatusTooManyRequests)
	}))
	data["b4"] = b4.URL

	// slow backend
	b5 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		<-time.After(time.Second)
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{"email": "some@email.com", "name": "a"})
	}))
	data["b5"] = b5.URL

	// querystring-forwarding backend
	b6 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		if ip := net.ParseIP(r.Header.Get("X-Forwarded-For")); ip == nil || !ip.IsLoopback() {
			http.Error(rw, "invalid X-Forwarded-For", 400)
			return
		}
		r.Header.Del("X-Forwarded-For")
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"path":    r.URL.Path,
			"query":   r.URL.Query(),
			"headers": r.Header,
		})
	}))
	data["b6"] = b6.URL

	// path validator
	b7 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{"path": r.URL.Path, "random": 42})
	}))
	data["b7"] = b7.URL

	// redirect
	b8 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, b7.URL, http.StatusMovedPermanently)
	}))
	data["b8"] = b8.URL

	// found
	b9 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Redirect(rw, r, b7.URL, http.StatusFound)
	}))
	data["b9"] = b9.URL

	// X-Forwarded-For backend
	b11 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{
			"headers": r.Header,
		})
	}))
	data["b11"] = b11.URL

	c, err := loadConfig(data)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func loadConfig(data map[string]interface{}) (*config.ServiceConfig, error) {
	content, _ := ioutil.ReadFile("lura.json")
	tmpl, err := template.New("test").Parse(string(content))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, data); err != nil {
		return nil, err
	}

	c, err := config.NewParserWithFileReader(func(s string) ([]byte, error) {
		return []byte(s), nil
	}).Parse(buf.String())

	if err != nil {
		return nil, err
	}

	return &c, nil
}
