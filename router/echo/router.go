/* Package httptreemux provides some basic implementations for building routers based on dimfeld/httptreemux
 */
// SPDX-License-Identifier: Apache-2.0
package httptreemux

import (
	"net/http"
	"strings"

	"github.com/dimfeld/httptreemux"
	"github.com/labstack/echo"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
	"github.com/luraproject/lura/router"
	"github.com/luraproject/lura/router/mux"
)

// DefaultFactory returns a net/http mux router factory with the injected proxy factory and logger
func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return mux.NewFactory(DefaultConfig(pf, logger))
}

// DefaultConfig returns the struct that collects the parts the router should be built from
func DefaultConfig(pf proxy.Factory, logger logging.Logger) mux.Config {
	return mux.Config{
		Engine:         &echoEngine{echo.New().Router()},
		Middlewares:    []mux.HandlerMiddleware{},
		HandlerFactory: mux.CustomEndpointHandler(mux.NewRequestBuilder(ParamsExtractor)),
		ProxyFactory:   pf,
		Logger:         logger,
		DebugPattern:   "/__debug/{params}",
		RunServer:      router.RunServer,
	}
}

func ParamsExtractor(r *http.Request) map[string]string {
	params := map[string]string{}
	for key, value := range httptreemux.ContextParams(r.Context()) {
		params[strings.Title(key)] = value
	}
	return params
}

type echoEngine struct {
	r *echo.Router
}

// Handle implements the mux.Engine interface from the lura router package
func (e echoEngine) Handle(pattern, method string, handler http.Handler) {
	e.r.Add(method, pattern, echo.WrapHandler(handler))
}

// ServeHTTP implements the http:Handler interface from the stdlib
func (e echoEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.r.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}
