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

	if cfg.Debug {
		r.registerDebugEndpoints()
	}

	r.registerKrakendEndpoints(cfg.Endpoints)

	r.cfg.Logger.Critical(r.cfg.Engine.Run(fmt.Sprintf(":%d", cfg.Port)))

}

func (r ginRouter) registerDebugEndpoints() {
	handler := DebugHandler(r.cfg.Logger)
	r.cfg.Engine.GET("/__debug/*param", handler)
	r.cfg.Engine.POST("/__debug/*param", handler)
	r.cfg.Engine.PUT("/__debug/*param", handler)
}

func (r ginRouter) registerKrakendEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}

		r.registerKrakendEndpoint(c.Method, c.Endpoint, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r ginRouter) registerKrakendEndpoint(method, path string, handler gin.HandlerFunc, totBackends int) {
	if method != "GET" && totBackends > 1 {
		r.cfg.Logger.Error(method, "endpoints must have a single backend! Ignoring", path)
		return
	}
	switch method {
	case "GET":
		r.cfg.Engine.GET(path, handler)
	case "POST":
		r.cfg.Engine.POST(path, handler)
	case "PUT":
		r.cfg.Engine.PUT(path, handler)
	case "PATCH":
		r.cfg.Engine.PATCH(path, handler)
	case "DELETE":
		r.cfg.Engine.DELETE(path, handler)
	default:
		r.cfg.Logger.Error("Unsupported method", method)
	}
}
