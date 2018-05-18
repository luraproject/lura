// Package mux provides some basic implementations for building routers based on net/http mux
package mux

import (
	"context"
	"fmt"
	"net/http"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
)

// DefaultDebugPattern is the default pattern used to define the debug endpoint
const DefaultDebugPattern = "/__debug/"

// Config is the struct that collects the parts the router should be builded from
type Config struct {
	Engine         Engine
	Middlewares    []HandlerMiddleware
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
	DebugPattern   string
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
	return httpRouter{rf.cfg, context.Background()}
}

// NewWithContext implements the factory interface
func (rf factory) NewWithContext(ctx context.Context) router.Router {
	return httpRouter{rf.cfg, ctx}
}

type httpRouter struct {
	cfg Config
	ctx context.Context
}

// Run implements the router interface
func (r httpRouter) Run(cfg config.ServiceConfig) {
	if cfg.Debug {
		r.cfg.Engine.Handle(r.cfg.DebugPattern, DebugHandler(r.cfg.Logger))
	}

	router.InitHTTPDefaultTransport(cfg)

	r.registerKrakendEndpoints(cfg.Endpoints)

	server := http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           r.handler(),
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	go func() {
		r.cfg.Logger.Critical(server.ListenAndServe())
	}()

	<-r.ctx.Done()
	if err := server.Shutdown(context.Background()); err != nil {
		r.cfg.Logger.Error(err.Error())
	}
	r.cfg.Logger.Info("Router execution ended")
}

func (r httpRouter) registerKrakendEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}

		r.registerKrakendEndpoint(c.Method, c.Endpoint, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r httpRouter) registerKrakendEndpoint(method, path string, handler http.HandlerFunc, totBackends int) {
	if method != "GET" && totBackends > 1 {
		r.cfg.Logger.Error(method, "endpoints must have a single backend! Ignoring", path)
		return
	}

	switch method {
	case "GET":
	case "POST":
	case "PUT":
	case "PATCH":
	case "DELETE":
	default:
		r.cfg.Logger.Error("Unsupported method", method)
		return
	}
	r.cfg.Logger.Debug("registering the endpoint", method, path)
	r.cfg.Engine.Handle(path, handler)
}

func (r httpRouter) handler() http.Handler {
	var handler http.Handler = r.cfg.Engine
	for _, middleware := range r.cfg.Middlewares {
		r.cfg.Logger.Debug("Adding the middleware", middleware)
		handler = middleware.Handler(handler)
	}
	return handler
}
