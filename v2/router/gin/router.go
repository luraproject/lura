/* Package gin provides some basic implementations for building routers based on gin-gonic/gin
 */
// SPDX-License-Identifier: Apache-2.0
package gin

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/router"
	"github.com/luraproject/lura/v2/transport/http/server"
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
			RunServer:      server.RunServer,
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
	return ginRouter{
		cfg:        rf.cfg,
		ctx:        ctx,
		runServerF: rf.cfg.RunServer,
		mu:         new(sync.Mutex),
	}
}

type ginRouter struct {
	cfg        Config
	ctx        context.Context
	runServerF RunServerFunc
	mu         *sync.Mutex
}

// Run implements the router interface
func (r ginRouter) Run(cfg config.ServiceConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		r.cfg.Logger.Debug("Debug enabled")
	}

	server.InitHTTPDefaultTransport(cfg)

	r.cfg.Engine.RedirectTrailingSlash = true
	r.cfg.Engine.RedirectFixedPath = true
	r.cfg.Engine.HandleMethodNotAllowed = true

	if v, ok := cfg.ExtraConfig[Namespace]; ok {
		b, err := json.Marshal(v)
		if err != nil {
			ginOptions := engineConfiguration{}
			if err := json.Unmarshal(b, &ginOptions); err == nil {
				r.cfg.Engine.RedirectTrailingSlash = !ginOptions.DisableRedirectTrailingSlash
				r.cfg.Engine.RedirectFixedPath = !ginOptions.DisableRedirectFixedPath
				r.cfg.Engine.HandleMethodNotAllowed = !ginOptions.DisableHandleMethodNotAllowed
				r.cfg.Engine.ForwardedByClientIP = ginOptions.ForwardedByClientIP
				r.cfg.Engine.RemoteIPHeaders = ginOptions.RemoteIPHeaders
				r.cfg.Engine.TrustedProxies = ginOptions.TrustedProxies
				r.cfg.Engine.AppEngine = ginOptions.AppEngine
				r.cfg.Engine.MaxMultipartMemory = ginOptions.MaxMultipartMemory
				r.cfg.Engine.RemoveExtraSlash = ginOptions.RemoveExtraSlash
			}
		}
	}

	if cfg.Debug {
		r.cfg.Engine.Any("/__debug/*param", DebugHandler(r.cfg.Logger))
	}

	r.cfg.Engine.GET("/__health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	endpointGroup := r.cfg.Engine.Group("/")
	endpointGroup.Use(r.cfg.Middlewares...)

	r.registerKrakendEndpoints(endpointGroup, cfg.Endpoints)

	if err := r.runServerF(r.ctx, cfg, r.cfg.Engine); err != nil {
		r.cfg.Logger.Error(err.Error())
	}

	r.cfg.Logger.Info("Router execution ended")
}

func (r ginRouter) registerKrakendEndpoints(rg *gin.RouterGroup, endpoints []*config.EndpointConfig) {
	for _, c := range endpoints {
		proxyStack, err := r.cfg.ProxyFactory.New(c)
		if err != nil {
			r.cfg.Logger.Error("calling the ProxyFactory", err.Error())
			continue
		}

		r.registerKrakendEndpoint(rg, c.Method, c, r.cfg.HandlerFactory(c, proxyStack), len(c.Backend))
	}
}

func (r ginRouter) registerKrakendEndpoint(rg *gin.RouterGroup, method string, e *config.EndpointConfig, h gin.HandlerFunc, total int) {
	method = strings.ToTitle(method)
	path := e.Endpoint
	if method != http.MethodGet && total > 1 {
		if !router.IsValidSequentialEndpoint(e) {
			r.cfg.Logger.Error(method, " endpoints with sequential enabled is only the last one is allowed to be non GET! Ignoring", path)
			return
		}
	}

	switch method {
	case http.MethodGet:
		rg.GET(path, h)
	case http.MethodPost:
		rg.POST(path, h)
	case http.MethodPut:
		rg.PUT(path, h)
	case http.MethodPatch:
		rg.PATCH(path, h)
	case http.MethodDelete:
		rg.DELETE(path, h)
	default:
		r.cfg.Logger.Error("Unsupported method", method)
	}
}

const Namespace = "github.com/luraproject/lura/router/gin"

func initEngine(e *gin.Engine, cfg map[string]interface{}) {
	raw, ok := cfg[Namespace]
	if !ok {
		return
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return
	}

	engineCfg := engineConfiguration{}
	if err := json.Unmarshal(b, &engineCfg); err != nil {
		return
	}

	e.RedirectTrailingSlash = true
	e.RedirectFixedPath = true
	e.HandleMethodNotAllowed = true
}

type engineConfiguration struct {
	// Disables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	DisableRedirectTrailingSlash bool `json:"disable_redirect_trailing_slash"`

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	DisableRedirectFixedPath bool `json:"disable_redirect_fixed_path"`

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	DisableHandleMethodNotAllowed bool `json:"disable_handle_method_not_allowed"`

	// If enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	ForwardedByClientIP bool `json:"forwarded_by_client_ip"`

	// List of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of `(*gin.Engine).TrustedProxies`.
	RemoteIPHeaders []string `json:"remote_ip_headers"`

	// List of network origins (IPv4 addresses, IPv4 CIDRs, IPv6 addresses or
	// IPv6 CIDRs) from which to trust request's headers that contain
	// alternative client IP when `(*gin.Engine).ForwardedByClientIP` is
	// `true`.
	TrustedProxies []string `json:"trusted_proxies"`

	// #726 #755 If enabled, it will trust some headers starting with
	// 'X-AppEngine...' for better integration with that PaaS.
	AppEngine bool `json:"app_engine"`

	/*
		// If enabled, the url.RawPath will be used to find parameters.
		UseRawPath bool `json:"use_raw_path"`

		// If true, the path value will be unescaped.
		// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
		// as url.Path gonna be used, which is already unescaped.
		UnescapePathValues bool `json:"unescape_path_values"`
	*/

	// Value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64 `json:"max_multipart_memory"`

	// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool `json:"remove_extra_slash"`
}
