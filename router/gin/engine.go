// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/transport/http/server"
)

const Namespace = "github_com/luraproject/lura/router/gin"

type EngineOptions struct {
	Logger    logging.Logger
	Writer    io.Writer
	Formatter gin.LogFormatter
	Health    <-chan string
}

// NewEngine returns an initialized gin engine
func NewEngine(cfg config.ServiceConfig, opt EngineOptions) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	if cfg.Debug {
		opt.Logger.Debug(logPrefix, "Debug enabled")
	}
	engine := gin.New()

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	paths := []string{}

	ginOptions := engineConfiguration{}
	if v, ok := cfg.ExtraConfig[Namespace]; ok {
		if b, err := json.Marshal(v); err == nil {
			if err := json.Unmarshal(b, &ginOptions); err == nil {
				engine.RedirectTrailingSlash = !ginOptions.DisableRedirectTrailingSlash
				engine.RedirectFixedPath = !ginOptions.DisableRedirectFixedPath
				engine.HandleMethodNotAllowed = !ginOptions.DisableHandleMethodNotAllowed
				engine.ForwardedByClientIP = ginOptions.ForwardedByClientIP
				engine.RemoteIPHeaders = ginOptions.RemoteIPHeaders
				for k, h := range engine.RemoteIPHeaders {
					engine.RemoteIPHeaders[k] = textproto.CanonicalMIMEHeaderKey(h)
				}
				engine.SetTrustedProxies(ginOptions.TrustedProxies)
				engine.AppEngine = ginOptions.AppEngine
				engine.MaxMultipartMemory = ginOptions.MaxMultipartMemory
				engine.RemoveExtraSlash = ginOptions.RemoveExtraSlash
				paths = ginOptions.LoggerSkipPaths

				returnErrorMsg = ginOptions.ReturnErrorMsg
			}
		}
	}

	engine.NoRoute(func(c *gin.Context) {
		c.Header(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
	})

	if !ginOptions.DisableAccessLog {
		engine.Use(
			gin.LoggerWithConfig(gin.LoggerConfig{
				Output:    opt.Writer,
				SkipPaths: paths,
				Formatter: opt.Formatter,
			}),
		)
	}
	engine.Use(gin.Recovery())

	if !ginOptions.DisablePathDecoding {
		engine.Use(paramChecker())
	}

	if !ginOptions.DisableHealthEndpoint {
		path := "/__health"
		if ginOptions.HealthPath != "" {
			path = ginOptions.HealthPath
		}

		engine.GET(path, healthEndpoint(opt.Health))
	}

	return engine
}

func healthEndpoint(health <-chan string) func(*gin.Context) {
	mu := new(sync.RWMutex)
	reports := map[string]string{}

	go func() {
		for name := range health {
			mu.Lock()
			reports[name] = time.Now().String()
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		mu.RLock()
		defer mu.RUnlock()

		c.JSON(200, gin.H{"status": "ok", "agents": reports, "now": time.Now().String()})
	}
}

func paramChecker() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, param := range c.Params {
			s, err := url.PathUnescape(param.Value)
			if err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("error: %s", err))
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			if s != param.Value {
				c.String(http.StatusBadRequest, "error: encoded url params")
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
		}
	}
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

	// Value of 'maxMemory' param that is given to http.Request's ParseMultipartForm
	// method call.
	MaxMultipartMemory int64 `json:"max_multipart_memory"`

	// RemoveExtraSlash a parameter can be parsed from the URL even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool `json:"remove_extra_slash"`

	// LoggerSkipPaths defines the set of path to avoid logging
	LoggerSkipPaths []string `json:"logger_skip_paths"`

	// AutoOptions enables the autogenerated OPTIONS endpoint for all the registered paths
	AutoOptions bool `json:"auto_options"`

	// ReturnErrorMsg flags if the error msg should be returned to the client as response body
	ReturnErrorMsg bool `json:"return_error_msg"`

	// DisableHealthEndpoint marks if the health check endpoint should be exposed
	DisableHealthEndpoint bool `json:"disable_health"`

	// HealthPath allows users to define a custom path for the health check endpoint
	HealthPath string `json:"health_path"`

	// DisableAccessLog blocks the injection of the router logger
	DisableAccessLog bool `json:"disable_access_log"`

	// Disables automatic validation of the url params looking for url encoded ones.
	// For example if /foo/..%252Fbar is requested and this flag is set to false, the router will
	// reject the request with http status code 400.
	DisablePathDecoding bool `json:"disable_path_decoding"`
}

var returnErrorMsg bool
