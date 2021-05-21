/* Package gorilla provides some basic implementations for building routers based on gorilla/mux
 */
// SPDX-License-Identifier: Apache-2.0
package gorilla

import (
	"net/http"
	"strings"

	gorilla "github.com/gorilla/mux"

	"github.com/luraproject/lura/logging"
	"github.com/luraproject/lura/proxy"
	"github.com/luraproject/lura/router"
	"github.com/luraproject/lura/router/mux"
)

// DefaultFactory returns a net/http mux router factory with the injected proxy factory and logger
func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return mux.NewFactory(DefaultConfig(pf, logger))
}

// DefaultConfig returns the struct that collects the parts the router should be builded from
func DefaultConfig(pf proxy.Factory, logger logging.Logger) mux.Config {
	return mux.Config{
		Engine:         gorillaEngine{gorilla.NewRouter()},
		Middlewares:    []mux.HandlerMiddleware{},
		HandlerFactory: mux.CustomEndpointHandler(mux.NewRequestBuilder(gorillaParamsExtractor)),
		ProxyFactory:   pf,
		Logger:         logger,
		DebugPattern:   "/__debug/{params}",
		RunServer:      router.RunServer,
	}
}

func gorillaParamsExtractor(r *http.Request) map[string]string {
	params := map[string]string{}
	for key, value := range gorilla.Vars(r) {
		params[strings.Title(key)] = value
	}
	return params
}

type gorillaEngine struct {
	r *gorilla.Router
}

// Handle implements the mux.Engine interface from the lura router package
func (g gorillaEngine) Handle(pattern, method string, handler http.Handler) {
	g.r.Handle(pattern, handler).Methods(method)
}

// ServeHTTP implements the http:Handler interface from the stdlib
func (g gorillaEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.r.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}
