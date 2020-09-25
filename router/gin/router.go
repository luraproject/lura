// Package gin provides some basic implementations for building routers based on gin-gonic/gin
package gin

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
)

// RunServerFunc is a func that will run the http Server with the given params.
type RunServerFunc func(context.Context, config.ServiceConfig, http.Handler) error

// Config is the struct that collects the parts the router should be builded from
type Config struct {
	Engine         *gin.Engine
	Middlewares    []gin.HandlerFunc
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
	RunServer      RunServerFunc
}

// DefaultFactory returns a gin router factory with the injected proxy factory and logger.
// It also uses a default gin router and the default HandlerFactory
func DefaultFactory(proxyFactory proxy.Factory, logger logging.Logger) router.Factory {
	return NewFactory(
		Config{
			Engine:         gin.Default(),
			Middlewares:    []gin.HandlerFunc{},
			HandlerFactory: EndpointHandler,
			ProxyFactory:   proxyFactory,
			Logger:         logger,
			RunServer:      router.RunServer,
		},
	)
}

// NewFactory returns a gin router factory with the injected configuration
func NewFactory(cfg Config) router.Factory {
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
	return ginRouter{rf.cfg, ctx, rf.cfg.RunServer}
}

type ginRouter struct {
	cfg       Config
	ctx       context.Context
	RunServer RunServerFunc
}

// Run implements the router interface
func (r ginRouter) Run(cfg config.ServiceConfig) {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		r.cfg.Logger.Debug("Debug enabled")
	}

	router.InitHTTPDefaultTransport(cfg)

	r.cfg.Engine.RedirectTrailingSlash = true
	r.cfg.Engine.RedirectFixedPath = true
	r.cfg.Engine.HandleMethodNotAllowed = true

	r.cfg.Engine.Use(r.cfg.Middlewares...)

	if cfg.Debug {
		r.cfg.Engine.Any("/__debug/*param", DebugHandler(r.cfg.Logger))
	}

	r.cfg.Engine.GET("/__health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.registerKrakendEndpoints(cfg.Endpoints)

	r.cfg.Engine.NoRoute(func(c *gin.Context) {
		c.Header(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
	})

	if err := r.RunServer(r.ctx, cfg, r.cfg.Engine); err != nil {
		r.cfg.Logger.Error(err.Error())
	}

	r.cfg.Logger.Info("Router execution ended")
}

func (r ginRouter) registerKrakendEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}

		r.registerKrakendEndpoint(c.Method, c, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r ginRouter) registerKrakendEndpoint(method string, endpoint *config.EndpointConfig, handler gin.HandlerFunc, totBackends int) {
	method = strings.ToTitle(method)
	path := endpoint.Endpoint
	if method != http.MethodGet && totBackends > 1 {
		if !router.IsValidSequentialEndpoint(endpoint) {
			r.cfg.Logger.Error(method, " endpoints with sequential enabled is only the last one is allowed to be non GET! Ignoring", path)
			return
		}
	}

	switch method {
	case http.MethodGet:
		r.cfg.Engine.GET(path, handler)
	case http.MethodPost:
		r.cfg.Engine.POST(path, handler)
	case http.MethodPut:
		r.cfg.Engine.PUT(path, handler)
	case http.MethodPatch:
		r.cfg.Engine.PATCH(path, handler)
	case http.MethodDelete:
		r.cfg.Engine.DELETE(path, handler)
	default:
		r.cfg.Logger.Error("Unsupported method", method)
	}
}
