/* Package echo provides some basic implementations for building routers based on labstack/echo
 */
// SPDX-License-Identifier: Apache-2.0
package echo

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
	"github.com/luraproject/lura/router"
	"github.com/luraproject/lura/router/mux"
)

const (
	textPlainContentType = "text/plain; charset=utf-8"
	pageNotFoundError    = "404 page not found\n"
)

// DefaultFactory returns a net/http mux router factory with the injected proxy factory and logger
func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return mux.NewFactory(DefaultConfig(pf, logger))
}

// DefaultConfig returns the struct that collects the parts the router should be built from
func DefaultConfig(pf proxy.Factory, logger logging.Logger) mux.Config {
	return mux.Config{
		Engine:         &engine{NewEchoEngine()},
		Middlewares:    []mux.HandlerMiddleware{},
		HandlerFactory: mux.CustomEndpointHandler(mux.NewRequestBuilder(echoParamsExtractor)),
		ProxyFactory:   pf,
		Logger:         logger,
		DebugPattern:   "/__debug/:params",
		RunServer:      router.RunServer,
	}
}

func echoParamsExtractor(r *http.Request) map[string]string {
	params := map[string]string{}
	keys := make([]string, 0, len(r.URL.Query()))
	for k := range r.URL.Query() {
		keys = append(keys, k)
	}

	for _, k := range keys {
		params[k] = r.URL.Query().Get(k)
	}

	return params
}

func NewEchoEngine() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = customHTTPErrorHandler

	return e
}

type engine struct {
	echo *echo.Echo
}

// Handle implements the mux.Engine interface from the lura router package
func (e engine) Handle(pattern, method string, handler http.Handler) {
	e.echo.Router().Add(method, pattern, echo.WrapHandler(handler))
}

// ServeHTTP implements the http:Handler interface from the stdlib
func (e engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.echo.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}

func customHTTPErrorHandler(err error, c echo.Context) {
	httpError, ok := err.(*echo.HTTPError)
	if ok {
		switch httpError.Code {
		case http.StatusNotFound:
			err := c.Blob(http.StatusNotFound, textPlainContentType, []byte(pageNotFoundError))
			c.Logger().Error(err)
		default:
			c.Logger().Error(httpError)
		}
	}
}
