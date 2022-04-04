// SPDX-License-Identifier: Apache-2.0

package mux

import (
	"net/http"
	"sync"

	"github.com/luraproject/lura/v2/transport/http/server"
)

// Engine defines the minimun required interface for the mux compatible engine
type Engine interface {
	http.Handler
	Handle(pattern, method string, handler http.Handler)
}

// BasicEngine is a slightly customized http.ServeMux router
type BasicEngine struct {
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
			i.ResponseWriter.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
		}
	})
	i.ResponseWriter.WriteHeader(code)
}

// DefaultEngine returns a new engine using BasicEngine
func DefaultEngine() *BasicEngine {
	return &BasicEngine{
		handler: http.NewServeMux(),
		dict:    map[string]map[string]http.HandlerFunc{},
	}
}

// Handle registers a handler at a given url pattern and http method
func (e *BasicEngine) Handle(pattern, method string, handler http.Handler) {
	if _, ok := e.dict[pattern]; !ok {
		e.dict[pattern] = map[string]http.HandlerFunc{}
		e.handler.Handle(pattern, e.registrableHandler(pattern))
	}
	e.dict[pattern][method] = handler.ServeHTTP
}

// ServeHTTP adds a error interceptor and delegates the request dispatching to the
// internal request multiplexer.
func (e *BasicEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.handler.ServeHTTP(NewHTTPErrorInterceptor(w), r)
}

func (e *BasicEngine) registrableHandler(pattern string) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if handler, ok := e.dict[pattern][req.Method]; ok {
			handler(rw, req)
			return
		}

		rw.Header().Set(server.CompleteResponseHeaderName, server.HeaderIncompleteResponseValue)
		http.Error(rw, "", http.StatusMethodNotAllowed)
	})
}
