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
	Handle(pattern string, handler http.Handler)
}

// DefaultEngine returns a new engine using a sligthly customized http.ServeMux router
func DefaultEngine() *engine {
	return &engine{http.NewServeMux()}
}

type engine struct {
	handler Engine
}

func (e *engine) Handle(pattern string, handler http.Handler) {
	e.handler.Handle(pattern, handler)
}

func (e *engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(NewHTTPErrorInterceptor(w), r)
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
