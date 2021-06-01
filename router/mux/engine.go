// Package mux provides some basic implementations for building routers based on net/http mux
// SPDX-License-Identifier: Apache-2.0
package mux

import (
	"net/http"
	"sync"

	"github.com/luraproject/lura/router"
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

// NewHTTPErrorInterceptor returns a HTTPErrorInterceptor over theinjected response writer
func NewHTTPErrorInterceptor(w http.ResponseWriter) *HTTPErrorInterceptor {
	return &HTTPErrorInterceptor{w, new(sync.Once)}
}

// HTTPErrorInterceptor is a reposnse writer that adds a header signaling incomplete response in case of
// seeing a status code not equal to 200
type HTTPErrorInterceptor struct {
	http.ResponseWriter
	once *sync.Once
}

// WriteHeader records the status code and adds a header signaling incomplete responses
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

// Handle registers a handler at a given url pattern and http method
func (e *engine) Handle(pattern, method string, handler http.Handler) {
	if _, ok := e.dict[pattern]; !ok {
		e.dict[pattern] = map[string]http.HandlerFunc{}
		e.handler.Handle(pattern, e.registrableHandler(pattern))
	}
	e.dict[pattern][method] = handler.ServeHTTP
}

// ServeHTTP adds a error interceptor and delegates the request dispatching to the
// internal request multiplexer.
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
