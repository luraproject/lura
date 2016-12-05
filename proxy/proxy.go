//Package proxy provides proxy and proxy middleware interfaces and implementations.
package proxy

import (
	"context"
	"errors"

	"github.com/devopsfaith/krakend/config"
)

// Response is the entity returned by the proxy
type Response struct {
	Data       map[string]interface{}
	IsComplete bool
}

var (
	ErrNoBackends       = errors.New("all endpoints must have at least one backend")
	ErrTooManyBackends  = errors.New("too many backends for this proxy")
	ErrTooManyProxies   = errors.New("too many proxies for this proxy middleware")
	ErrNotEnoughProxies = errors.New("not enough proxies for this endpoint")
)

// Proxy processes a request in a given context and returns a response and an error
type Proxy func(ctx context.Context, request *Request) (*Response, error)

// BackendFactory creates a proxy based on the received backend configuration
type BackendFactory func(remote *config.Backend) Proxy

// Middleware adds a middleware, decorator or wrapper over a collection of proxies,
// exposing a proxy interface.
//
// Proxy middlewares can be stacked:
//	var p Proxy
//	p := EmptyMiddleware(NoopProxy)
//	response, err := p(ctx, r)
type Middleware func(next ...Proxy) Proxy

func EmptyMiddleware(next ...Proxy) Proxy {
	if len(next) > 1 {
		panic(ErrTooManyProxies)
	}
	return next[0]
}

func NoopProxy(_ context.Context, _ *Request) (*Response, error) { return nil, nil }
