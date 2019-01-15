package httptreemux

import (
	"net/http"

	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
	"github.com/devopsfaith/krakend/router/mux"
	"github.com/dimfeld/httptreemux"
)

// DefaultFactory returns a net/http mux router factory with the injected proxy factory and logger
func DefaultFactory(pf proxy.Factory, logger logging.Logger) router.Factory {
	return mux.NewFactory(DefaultConfig(pf, logger))
}

// DefaultConfig returns the struct that collects the parts the router should be built from
func DefaultConfig(pf proxy.Factory, logger logging.Logger) mux.Config {
	return mux.Config{
		Engine:         Engine{httptreemux.NewContextMux()},
		Middlewares:    []mux.HandlerMiddleware{},
		HandlerFactory: mux.CustomEndpointHandler(mux.NewRequestBuilder(ParamsExtractor)),
		ProxyFactory:   pf,
		Logger:         logger,
		DebugPattern:   "/__debug/{params}",
		RunServer:      router.RunServer,
	}
}

func ParamsExtractor(r *http.Request) map[string]string {
	return httptreemux.ContextParams(r.Context())
}

func NewEngine(m *httptreemux.ContextMux) Engine {
	return Engine{m}
}

type Engine struct {
	r *httptreemux.ContextMux
}

// Handle implements the mux.Engine interface from the krakend router package
func (g Engine) Handle(pattern, method string, handler http.Handler) {
	g.r.Handle(method, pattern, handler.(http.HandlerFunc))
}

// ServeHTTP implements the http:Handler interface from the stdlib
func (g Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.r.ServeHTTP(mux.NewHTTPErrorInterceptor(w), r)
}
