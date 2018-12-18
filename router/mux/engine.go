// Package mux provides some basic implementations for building routers based on net/http mux
package mux

import (
	"net/http"
	"sync"

	"github.com/devopsfaith/krakend/router"
)

// Engine defines the minimun required interface for the mux compatible engine
type Engine interface {
	http.Handler
	Handle(pattern, method string, handler http.Handler)
}

type engine struct {
	handler *http.ServeMux
	dict    map[string]map[string]http.HandlerFunc
}

func NewHTTPErrorInterceptor(w http.ResponseWriter) *HTTPErrorInterceptor {
	return &HTTPErrorInterceptor{w, new(sync.Once)}
}

type HTTPErrorInterceptor struct {
	http.ResponseWriter
	once *sync.Once
}

func (i *HTTPErrorInterceptor) WriteHeader(code int) {
	i.once.Do(func() {
		if code != http.StatusOK {
			i.ResponseWriter.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
		}
	})
	i.ResponseWriter.WriteHeader(code)
}

// DefaultEngine returns a new engine using a slightly customized http.ServeMux router
func DefaultEngine() *engine {
	return &engine{
		handler: http.NewServeMux(),
		dict:    map[string]map[string]http.HandlerFunc{},
	}
}

func (e *engine) Handle(pattern, method string, handler http.Handler) {
	if _, ok := e.dict[pattern]; !ok {
		e.dict[pattern] = map[string]http.HandlerFunc{}
		e.handler.Handle(pattern, e.registrableHandler(pattern))
	}
	e.dict[pattern][method] = handler.ServeHTTP
}

func (e *engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(NewHTTPErrorInterceptor(w), r)
}

func (e *engine) registrableHandler(pattern string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if handler, ok := e.dict[pattern][req.Method]; ok {
			handler(rw, req)
			return
		}

		rw.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
		http.Error(rw, "", http.StatusMethodNotAllowed)
	})
}
