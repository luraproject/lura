// SPDX-License-Identifier: Apache-2.0
package gin

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

const Namespace = "github.com/luraproject/lura/router/gin"

// NewEngine returns an initialized gin engine
func NewEngine(cfg config.ServiceConfig, logger logging.Logger, w io.Writer) *gin.Engine {
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	} else {
		logger.Debug("Debug enabled")
	}
	engine := gin.New()

	engine.RedirectTrailingSlash = true
	engine.RedirectFixedPath = true
	engine.HandleMethodNotAllowed = true

	paths := []string{}

	if v, ok := cfg.ExtraConfig[Namespace]; ok {
		b, err := json.Marshal(v)
		if err != nil {
			ginOptions := engineConfiguration{}
			if err := json.Unmarshal(b, &ginOptions); err == nil {
				engine.RedirectTrailingSlash = !ginOptions.DisableRedirectTrailingSlash
				engine.RedirectFixedPath = !ginOptions.DisableRedirectFixedPath
				engine.HandleMethodNotAllowed = !ginOptions.DisableHandleMethodNotAllowed
				engine.ForwardedByClientIP = ginOptions.ForwardedByClientIP
				engine.RemoteIPHeaders = ginOptions.RemoteIPHeaders
				engine.TrustedProxies = ginOptions.TrustedProxies
				engine.AppEngine = ginOptions.AppEngine
				engine.MaxMultipartMemory = ginOptions.MaxMultipartMemory
				engine.RemoveExtraSlash = ginOptions.RemoveExtraSlash
				paths = ginOptions.LoggerSkipPaths
			}
		}
	}

	engine.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{Output: w, SkipPaths: paths}),
		gin.Recovery(),
	)

	return engine
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

	// LoggerSkipPaths defines the set of path to avoid logging
	LoggerSkipPaths []string `json:"logger_skip_paths"`
}
