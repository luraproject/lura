// Package gin provides some basic implementations for building routers based on gin-gonic/gin
package gin

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
)

// Config is the struct that collects the parts the router should be builded from
type Config struct {
	Engine         *gin.Engine
	Middlewares    []gin.HandlerFunc
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
}

// DefaultFactory returns a gin router factory with the injected proxy factory and logger.
// It also uses a default gin router and the default HandlerFactory
func DefaultFactory(proxyFactory proxy.Factory, logger logging.Logger) router.Factory {
	return factory{
		Config{
			Engine:         gin.Default(),
			Middlewares:    []gin.HandlerFunc{},
			HandlerFactory: EndpointHandler,
			ProxyFactory:   proxyFactory,
			Logger:         logger,
		},
	}
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
	return ginRouter{rf.cfg}
}

type ginRouter struct {
	cfg Config
}

// Run implements the router interface
func (r ginRouter) Run(cfg config.ServiceConfig) {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		r.cfg.Logger.Debug("Debug enabled")
	}

	r.cfg.Engine.RedirectTrailingSlash = true
	r.cfg.Engine.RedirectFixedPath = true
	r.cfg.Engine.HandleMethodNotAllowed = true

	r.cfg.Engine.Use(r.cfg.Middlewares...)

	for _, c := range cfg.Endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}
		handler := r.cfg.HandlerFactory(c, proxyStack)

		switch c.Method {
		case "GET":
			r.cfg.Engine.GET(c.Endpoint, handler)
		case "POST":
			if len(c.Backend) > 1 {
				r.cfg.Logger.Error("POST endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
			r.cfg.Engine.POST(c.Endpoint, handler)
		case "PUT":
			if len(c.Backend) > 1 {
				r.cfg.Logger.Error("PUT endpoints must have a single backend! Ignoring", c.Endpoint)
				continue
			}
			r.cfg.Engine.PUT(c.Endpoint, handler)
		default:
			r.cfg.Logger.Error("Unsupported method", c.Method)
		}
	}

	if cfg.Debug {
		handler := DebugHandler(r.cfg.Logger)
		r.cfg.Engine.GET("/__debug/*param", handler)
		r.cfg.Engine.POST("/__debug/*param", handler)
		r.cfg.Engine.PUT("/__debug/*param", handler)
	}

	r.cfg.Logger.Critical(r.cfg.Engine.Run(fmt.Sprintf(":%d", cfg.Port)))

}
