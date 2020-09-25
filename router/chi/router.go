package chi

import (
	"context"
	"net/http"
	"strings"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
	"github.com/devopsfaith/krakend/router/mux"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// ChiDefaultDebugPattern is the default pattern used to define the debug endpoint
const ChiDefaultDebugPattern = "/__debug/"

// RunServerFunc is a func that will run the http Server with the given params.
type RunServerFunc func(context.Context, config.ServiceConfig, http.Handler) error

// Config is the struct that collects the parts the router should be builded from
type Config struct {
	Engine         chi.Router
	Middlewares    chi.Middlewares
	HandlerFactory HandlerFactory
	ProxyFactory   proxy.Factory
	Logger         logging.Logger
	DebugPattern   string
	RunServer      RunServerFunc
}

// DefaultFactory returns a chi router factory with the injected proxy factory and logger.
// It also uses a default chi router and the default HandlerFactory
func DefaultFactory(proxyFactory proxy.Factory, logger logging.Logger) router.Factory {
	return NewFactory(
		Config{
			Engine:         chi.NewRouter(),
			Middlewares:    chi.Middlewares{middleware.Logger},
			HandlerFactory: NewEndpointHandler,
			ProxyFactory:   proxyFactory,
			Logger:         logger,
			DebugPattern:   ChiDefaultDebugPattern,
			RunServer:      router.RunServer,
		},
	)
}

// NewFactory returns a chi router factory with the injected configuration
func NewFactory(cfg Config) router.Factory {
	if cfg.DebugPattern == "" {
		cfg.DebugPattern = ChiDefaultDebugPattern
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
	return chiRouter{rf.cfg, ctx, rf.cfg.RunServer}
}

type chiRouter struct {
	cfg       Config
	ctx       context.Context
	RunServer RunServerFunc
}

// Run implements the router interface
func (r chiRouter) Run(cfg config.ServiceConfig) {
	r.cfg.Engine.Use(r.cfg.Middlewares...)

	if cfg.Debug {
		r.registerDebugEndpoints()
	}

	r.cfg.Engine.Get("/__health", mux.HealthHandler)

	router.InitHTTPDefaultTransport(cfg)

	r.registerKrakendEndpoints(cfg.Endpoints)

	r.cfg.Engine.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
		http.NotFound(w, r)
	})

	if err := r.RunServer(r.ctx, cfg, r.cfg.Engine); err != nil {
		r.cfg.Logger.Error(err.Error())
	}

	r.cfg.Logger.Info("Router execution ended")
}

func (r chiRouter) registerDebugEndpoints() {
	debugHandler := mux.DebugHandler(r.cfg.Logger)
	r.cfg.Engine.Get(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Post(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Put(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Patch(r.cfg.DebugPattern, debugHandler)
	r.cfg.Engine.Delete(r.cfg.DebugPattern, debugHandler)
}

func (r chiRouter) registerKrakendEndpoints(endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}

		r.registerKrakendEndpoint(c.Method, c, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r chiRouter) registerKrakendEndpoint(method string, endpoint *config.EndpointConfig, handler http.HandlerFunc, totBackends int) {
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
		r.cfg.Engine.Get(path, handler)
	case http.MethodPost:
		r.cfg.Engine.Post(path, handler)
	case http.MethodPut:
		r.cfg.Engine.Put(path, handler)
	case http.MethodPatch:
		r.cfg.Engine.Patch(path, handler)
	case http.MethodDelete:
		r.cfg.Engine.Delete(path, handler)
	default:
		r.cfg.Logger.Error("Unsupported method", method)
		return
	}
	r.cfg.Logger.Debug("registering the endpoint", method, path)
}
