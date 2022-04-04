// SPDX-License-Identifier: Apache-2.0

/*
	Package mux provides some basic implementations for building routers based on net/http mux
*/
package mux

import (
	"context"
	"net/http"
	"strings"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/router"
	"github.com/luraproject/lura/v2/transport/http/server"
)

// DefaultDebugPattern is the default pattern used to define the debug endpoint
const DefaultDebugPattern = "/__debug/"
const logPrefix = "[SERVICE: Mux]"

// RunServerFunc is a func that will run the http Server with the given params.
type RunServerFunc func(context.Context, config.ServiceConfig, http.Handler) error

// Config is the struct that collects the parts the router should be builded from
type Config struct {
	Engine         Engine
	Middlewares    []HandlerMiddleware
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
	DebugPattern   string
	RunServer      RunServerFunc
}

// HandlerMiddleware is the interface for the decorators over the http.Handler
type HandlerMiddleware interface {
	Handler(h http.Handler) http.Handler
}

// DefaultFactory returns a net/http mux router factory with the injected proxy factory and logger
func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return factory{
		Config{
			Engine:         DefaultEngine(),
			Middlewares:    []HandlerMiddleware{},
			HandlerFactory: EndpointHandler,
			ProxyFactory:   pf,
			Logger:         logger,
			DebugPattern:   DefaultDebugPattern,
			RunServer:      server.RunServer,
		},
	}
}

// NewFactory returns a net/http mux router factory with the injected configuration
func NewFactory(cfg Config) router.Factory {
	if cfg.DebugPattern == "" {
		cfg.DebugPattern = DefaultDebugPattern
	}
	return factory{cfg}
}

type factory struct {
	cfg Config
}

// New implements the factory interface
func (rf factory) New() router.Router {
	return rf.NewWithContext(context.Background())
}

// NewWithContext implements the factory interface
func (rf factory) NewWithContext(ctx context.Context) router.Router {
	return httpRouter{rf.cfg, ctx, rf.cfg.RunServer}
}

type httpRouter struct {
	cfg       Config
	ctx       context.Context
	RunServer RunServerFunc
}

// HealthHandler is a dummy http.HandlerFunc implementation for exposing a health check endpoint
func HealthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// Run implements the router interface
func (r httpRouter) Run(cfg config.ServiceConfig) {
	if cfg.Debug {
		debugHandler := DebugHandler(r.cfg.Logger)
		for _, method := range []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
			http.MethodConnect,
			http.MethodTrace,
		} {
			r.cfg.Engine.Handle(r.cfg.DebugPattern, method, debugHandler)
		}
	}
	r.cfg.Engine.Handle("/__health", "GET", http.HandlerFunc(HealthHandler))

	server.InitHTTPDefaultTransport(cfg)

	r.registerKrakendEndpoints(cfg.Endpoints)

	if err := r.RunServer(r.ctx, cfg, r.handler()); err != nil {
		r.cfg.Logger.Error(logPrefix, err.Error())
	}

	r.cfg.Logger.Info(logPrefix, "Router execution ended")
}

func (r httpRouter) registerKrakendEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error(logPrefix, "Calling the ProxyFactory", err.Error())
			continue
		}

		r.registerKrakendEndpoint(c.Method, c, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r httpRouter) registerKrakendEndpoint(method string, endpoint *config.EndpointConfig, handler http.HandlerFunc, totBackends int) {
	method = strings.ToTitle(method)
	path := endpoint.Endpoint
	if method != http.MethodGet && totBackends > 1 {
		if !router.IsValidSequentialEndpoint(endpoint) {
			r.cfg.Logger.Error(logPrefix, method, " endpoints with sequential proxy enabled only allow a non-GET in the last backend! Ignoring", path)
			return
		}
	}

	switch method {
	case http.MethodGet:
	case http.MethodPost:
	case http.MethodPut:
	case http.MethodPatch:
	case http.MethodDelete:
	default:
		r.cfg.Logger.Error(logPrefix, "Unsupported method", method)
		return
	}
	r.cfg.Logger.Debug(logPrefix, "Registering the endpoint", method, path)
	r.cfg.Engine.Handle(path, method, handler)
}

func (r httpRouter) handler() http.Handler {
	var handler http.Handler = r.cfg.Engine
	for _, middleware := range r.cfg.Middlewares {
		r.cfg.Logger.Debug(logPrefix, "Adding the middleware", middleware)
		handler = middleware.Handler(handler)
	}
	return handler
}
