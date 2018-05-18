// Package router defines some interfaces for router adapters
package router

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
)

// Router sets up the public layer exposed to the users
type Router interface {
	Run(config.ServiceConfig)
}

// RouterFunc type is an adapter to allow the use of ordinary functions as routers.
// If f is a function with the appropriate signature, RouterFunc(f) is a Router that calls f.
type RouterFunc func(config.ServiceConfig)

// Run implements the Router interface
func (f RouterFunc) Run(cfg config.ServiceConfig) { f(cfg) }

// Factory creates new routers
type Factory interface {
	New() Router
	NewWithContext(context.Context) Router
}

// ToHTTPError translates an error into a HTTP status code
type ToHTTPError func(error) int

// DefaultToHTTPError is a ToHTTPError transalator that always returns an
// internal server error
func DefaultToHTTPError(_ error) int {
	return http.StatusInternalServerError
}

const (
	HeaderCompleteResponseValue   = "true"
	HeaderIncompleteResponseValue = "false"
)

var (
	// CompleteResponseHeaderName is the header to flag incomplete responses to the client
	CompleteResponseHeaderName = "X-KrakenD-Completed"
	// HeadersToSend are the headers to pass from the router request to the proxy
	HeadersToSend = []string{"Content-Type"}
	// UserAgentHeaderValue is the value of the User-Agent header to add to the proxy request
	UserAgentHeaderValue = []string{core.KrakendUserAgent}
	// ErrInternalError is the error returned by the router when something went wrong
	ErrInternalError = errors.New("internal server error")

	onceTransportConfig sync.Once
)

// InitHTTPDefaultTransport ensures the default HTTP transport is configured just once per execution
func InitHTTPDefaultTransport(cfg config.ServiceConfig) {
	onceTransportConfig.Do(func() {
		transport := http.DefaultTransport.(*http.Transport)
		transport.DisableCompression = cfg.DisableCompression
		transport.DisableKeepAlives = cfg.DisableKeepAlives
		transport.MaxIdleConns = cfg.MaxIdleConns
		transport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
		transport.IdleConnTimeout = cfg.IdleConnTimeout
		transport.ResponseHeaderTimeout = cfg.ResponseHeaderTimeout
		transport.ExpectContinueTimeout = cfg.ExpectContinueTimeout
	})
}
