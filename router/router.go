// Package router defines some interfaces for router adapters
package router

import (
	"context"

	"github.com/devopsfaith/krakend/config"
	http "github.com/devopsfaith/krakend/transport/http/server"
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
type ToHTTPError http.ToHTTPError

// DefaultToHTTPError is a ToHTTPError transalator that always returns an
// internal server error
var DefaultToHTTPError = http.DefaultToHTTPError

const (
	HeaderCompleteResponseValue   = http.HeaderCompleteResponseValue
	HeaderIncompleteResponseValue = http.HeaderIncompleteResponseValue
)

var (
	// CompleteResponseHeaderName is the header to flag incomplete responses to the client
	CompleteResponseHeaderName = http.CompleteResponseHeaderName
	// HeadersToSend are the headers to pass from the router request to the proxy
	HeadersToSend = http.HeadersToSend
	// UserAgentHeaderValue is the value of the User-Agent header to add to the proxy request
	UserAgentHeaderValue = http.UserAgentHeaderValue
	// ErrInternalError is the error returned by the router when something went wrong
	ErrInternalError = http.ErrInternalError
)

// InitHTTPDefaultTransport ensures the default HTTP transport is configured just once per execution
var InitHTTPDefaultTransport = http.InitHTTPDefaultTransport

// RunServer runs a http.Server with the given handler and configuration
var RunServer = http.RunServer
