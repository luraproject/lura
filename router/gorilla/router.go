package gorilla

import (
	"net/http"
	"strings"

	"github.com/devopsfaith/krakend/router"
	gorilla "github.com/gorilla/mux"

	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router/mux"
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

// Handle implements the mux.Engine interface from the krakend router package
func (g gorillaEngine) Handle(pattern string, handler http.Handler) {
	g.r.Handle(pattern, handler)
}

// ServeHTTP implements the http:Handler interface from the stdlib
func (g gorillaEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.r.ServeHTTP(w, r)
}
