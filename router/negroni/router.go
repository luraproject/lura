package negroni

import (
	"net/http"

	gorilla "github.com/gorilla/mux"
	"github.com/urfave/negroni"

	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
	krakendgorilla "github.com/devopsfaith/krakend/router/gorilla"
	"github.com/devopsfaith/krakend/router/mux"
)

// DefaultFactory returns a net/http mux router factory with the injected proxy factory and logger
func DefaultFactory(pf proxy.Factory, logger logging.Logger, middlewares []negroni.Handler) router.Factory {
	return mux.NewFactory(DefaultConfig(pf, logger, middlewares))
}

// DefaultConfig returns the struct that collects the parts the router should be builded from
func DefaultConfig(pf proxy.Factory, logger logging.Logger, middlewares []negroni.Handler) mux.Config {
	return DefaultConfigWithRouter(pf, logger, NewGorillaRouter(), middlewares)
}

// DefaultConfigWithRouter returns the struct that collects the parts the router should be builded from with the
// injected gorilla mux router
func DefaultConfigWithRouter(pf proxy.Factory, logger logging.Logger, muxEngine *gorilla.Router, middlewares []negroni.Handler) mux.Config {
	cfg := krakendgorilla.DefaultConfig(pf, logger)
	cfg.Engine = newNegroniEngine(muxEngine, middlewares...)
	return cfg
}

// NewGorillaRouter is a wrapper over the default gorilla router builder
func NewGorillaRouter() *gorilla.Router {
	return gorilla.NewRouter()
}

func newNegroniEngine(muxEngine *gorilla.Router, middlewares ...negroni.Handler) negroniEngine {
	negroniRouter := negroni.Classic()
	for _, m := range middlewares {
		negroniRouter.Use(m)
	}

	negroniRouter.UseHandler(muxEngine)

	return negroniEngine{muxEngine, negroniRouter}
}

type negroniEngine struct {
	r *gorilla.Router
	n *negroni.Negroni
}

// Handle implements the mux.Engine interface from the krakend router package
func (e negroniEngine) Handle(pattern string, handler http.Handler) {
	e.r.Handle(pattern, handler)
}

// ServeHTTP implements the http:Handler interface from the stdlib
func (e negroniEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.n.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}
