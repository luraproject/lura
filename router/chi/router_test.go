//go:build !race
// +build !race

// SPDX-License-Identifier: Apache-2.0

package chi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/transport/http/server"
)

func TestDefaultFactory_ok(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := logging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(5 * time.Millisecond)
	}()

	r := DefaultFactory(noopProxyFactory(map[string]interface{}{"supu": "tupu"}), logger).NewWithContext(ctx)
	expectedBody := "{\"supu\":\"tupu\"}"

	serviceCfg := config.ServiceConfig{
		Port:    8062,
		Version: 3,
		Host:    []string{"http://example.com"},
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/get",
				Method:   "GET",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/get",
				Method:   "POST",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/post",
				Method:   "Post",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/put",
				Method:   "put",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/patch",
				Method:   "PATCH",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/delete",
				Method:   "DELETE",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
		},
	}

	if err = serviceCfg.Init(); err != nil {
		t.Errorf("Error during the config init: %s\n", err.Error())
	}

	go func() { r.Run(serviceCfg) }()

	time.Sleep(5 * time.Millisecond)

	for _, endpoint := range serviceCfg.Endpoints {
		req, _ := http.NewRequest(strings.ToTitle(endpoint.Method), fmt.Sprintf("http://127.0.0.1:8062%s", endpoint.Endpoint), http.NoBody)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error("Making the request:", err.Error())
			return
		}
		defer resp.Body.Close()

		body, ioerr := io.ReadAll(resp.Body)
		if ioerr != nil {
			t.Error("Reading the response:", ioerr.Error())
			return
		}
		content := string(body)
		if resp.Header.Get("Cache-Control") != "" {
			t.Error(endpoint.Endpoint, "Cache-Control error:", resp.Header.Get("Cache-Control"))
		}
		if resp.Header.Get(server.CompleteResponseHeaderName) != server.HeaderCompleteResponseValue {
			t.Error(server.CompleteResponseHeaderName, "error:", resp.Header.Get(server.CompleteResponseHeaderName))
		}
		if resp.Header.Get("Content-Type") != "application/json" {
			t.Error(endpoint.Endpoint, "Content-Type error:", resp.Header.Get("Content-Type"))
		}
		if resp.Header.Get("X-Krakend") != "Version undefined" {
			t.Error(endpoint.Endpoint, "X-Krakend error:", resp.Header.Get("X-Krakend"))
		}
		if resp.StatusCode != http.StatusOK {
			t.Error(endpoint.Endpoint, "Unexpected status code:", resp.StatusCode)
		}
		if content != expectedBody {
			t.Error(endpoint.Endpoint, "Unexpected body:", content, "expected:", expectedBody)
		}
	}
}

func TestDefaultFactory_ko(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := logging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(5 * time.Millisecond)
	}()

	r := NewFactory(Config{
		Engine:         chi.NewRouter(),
		Middlewares:    chi.Middlewares{},
		HandlerFactory: NewEndpointHandler,
		ProxyFactory:   noopProxyFactory(map[string]interface{}{"supu": "tupu"}),
		Logger:         logger,
		RunServer:      server.RunServer,
	}).NewWithContext(ctx)

	serviceCfg := config.ServiceConfig{
		Debug: true,
		Port:  8063,
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/ignored",
				Method:   "GETTT",
				Backend: []*config.Backend{
					{},
				},
			},
			{
				Endpoint: "/empty",
				Method:   "GETTT",
				Backend:  []*config.Backend{},
			},
			{
				Endpoint: "/no-hosts-ignored",
				Method:   "GET",
				Backend: []*config.Backend{
					{Host: []string{}},
				},
			},
		},
	}

	go func() { r.Run(serviceCfg) }()

	time.Sleep(5 * time.Millisecond)

	for _, subject := range [][]string{
		{"GET", "ignored"},
		{"GET", "empty"},
		{"GET", "no-hosts-ignored"},
	} {
		req, _ := http.NewRequest(subject[0], fmt.Sprintf("http://127.0.0.1:8063/%s", subject[1]), http.NoBody)
		req.Header.Set("Content-Type", "application/json")
		checkResponseIs404(t, req)
	}
}

func TestDefaultFactory_proxyFactoryCrash(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := logging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		time.Sleep(5 * time.Millisecond)
	}()

	r := DefaultFactory(erroredProxyFactory{fmt.Errorf("%s", "crash!!!")}, logger).NewWithContext(ctx)

	serviceCfg := config.ServiceConfig{
		Debug: true,
		Port:  8064,
		Endpoints: []*config.EndpointConfig{
			{
				Endpoint: "/ignored",
				Method:   "GET",
				Timeout:  10,
				Backend: []*config.Backend{
					{},
				},
			},
		},
	}

	go func() { r.Run(serviceCfg) }()

	time.Sleep(5 * time.Millisecond)

	for _, subject := range [][]string{{"GET", "ignored"}, {"PUT", "also-ignored"}} {
		req, _ := http.NewRequest(subject[0], fmt.Sprintf("http://127.0.0.1:8064/%s", subject[1]), http.NoBody)
		req.Header.Set("Content-Type", "application/json")
		checkResponseIs404(t, req)
	}
}

func TestRunServer_ko(t *testing.T) {
	buff := new(bytes.Buffer)
	logger, err := logging.NewLogger("DEBUG", buff, "")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	errorMsg := "runServer error"
	runServerFunc := func(_ context.Context, _ config.ServiceConfig, _ http.Handler) error {
		return errors.New(errorMsg)
	}

	pf := noopProxyFactory(map[string]interface{}{"supu": "tupu"})
	r := NewFactory(
		Config{
			Engine:         chi.NewRouter(),
			Middlewares:    chi.Middlewares{},
			HandlerFactory: NewEndpointHandler,
			ProxyFactory:   pf,
			Logger:         logger,
			DebugPattern:   ChiDefaultDebugPattern,
			RunServer:      runServerFunc,
		},
	).New()

	serviceCfg := config.ServiceConfig{}
	r.Run(serviceCfg)
	re := regexp.MustCompile(errorMsg)
	if !re.MatchString(buff.String()) {
		t.Errorf("the logger doesn't contain the expected msg: %s", buff.Bytes())
	}
}

func checkResponseIs404(t *testing.T, req *http.Request) {
	expectedBody := "404 page not found\n"
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("Making the request:", err.Error())
		return
	}
	defer resp.Body.Close()
	body, ioerr := io.ReadAll(resp.Body)
	if ioerr != nil {
		t.Error("Reading the response:", ioerr.Error())
		return
	}
	content := string(body)

	if resp.Header.Get("Cache-Control") != "" {
		t.Error("Cache-Control error:", resp.Header.Get("Cache-Control"))
	}
	if resp.Header.Get(server.CompleteResponseHeaderName) != server.HeaderIncompleteResponseValue {
		t.Error(req.URL.String(), server.CompleteResponseHeaderName, "error:", resp.Header.Get(server.CompleteResponseHeaderName))
	}
	if resp.Header.Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Error("Content-Type error:", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("X-Krakend") != "" {
		t.Error("X-Krakend error:", resp.Header.Get("X-Krakend"))
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Error("Unexpected status code:", resp.StatusCode)
	}
	if content != expectedBody {
		t.Error("Unexpected body:", content, "expected:", expectedBody)
	}
}

type noopProxyFactory map[string]interface{}

func (n noopProxyFactory) New(_ *config.EndpointConfig) (proxy.Proxy, error) {
	return func(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
		return &proxy.Response{
			IsComplete: true,
			Data:       n,
		}, nil
	}, nil
}

type erroredProxyFactory struct {
	Error error
}

func (e erroredProxyFactory) New(_ *config.EndpointConfig) (proxy.Proxy, error) {
	return proxy.NoopProxy, e.Error
}
