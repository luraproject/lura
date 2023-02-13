// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"time"

	"github.com/luraproject/lura/v2/config"
)

const (
	shadowKey = "shadow"
)

type shadowFactory struct {
	f Factory
}

// New check the Backends for an ExtraConfig with the "shadow" param to true
// implements the Factory interface
func (s shadowFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	if len(cfg.Backend) == 0 {
		err = ErrNoBackends
		return
	}

	var shadow []*config.Backend

	var regular []*config.Backend

	for _, b := range cfg.Backend {
		if isShadowBackend(b) {
			shadow = append(shadow, b)
			continue
		}
		regular = append(regular, b)
	}

	cfg.Backend = regular

	p, err = s.f.New(cfg)

	if len(shadow) > 0 {
		cfg.Backend = shadow
		pShadow, _ := s.f.New(cfg)
		p = ShadowMiddlewareWithTimeout(cfg.Timeout, p, pShadow)
	}

	return
}

// NewShadowFactory creates a new shadowFactory using the provided Factory
func NewShadowFactory(f Factory) Factory {
	return shadowFactory{f}
}

// ShadowMiddleware is a Middleware that creates a shadowProxy
func ShadowMiddleware(next ...Proxy) Proxy {
	switch len(next) {
	case 0:
		panic(ErrNotEnoughProxies)
	case 1:
		return next[0]
	case 2:
		return NewShadowProxy(next[0], next[1])
	default:
		panic(ErrTooManyProxies)
	}
}

func ShadowMiddlewareWithTimeout(timeout time.Duration, next ...Proxy) Proxy {
	switch len(next) {
	case 0:
		panic(ErrNotEnoughProxies)
	case 1:
		return next[0]
	case 2:
		return NewShadowProxyWithTimeout(timeout, next[0], next[1])
	default:
		panic(ErrTooManyProxies)
	}
}

// NewShadowProxy returns a Proxy that sends requests to p1 and p2 but ignores
// the response of p2
func NewShadowProxy(p1, p2 Proxy) Proxy {
	return NewShadowProxyWithTimeout(config.DefaultTimeout, p1, p2)
}

func NewShadowProxyWithTimeout(timeout time.Duration, p1, p2 Proxy) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		shadowCtx, cancel := newcontextWrapperWithTimeout(ctx, timeout)
		go func() {
			p2(shadowCtx, CloneRequest(request))
			cancel()
		}()
		return p1(ctx, request)
	}
}

func isShadowBackend(c *config.Backend) bool {
	if v, ok := c.ExtraConfig[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if v, ok := e[shadowKey]; ok {
				c, ok := v.(bool)
				return ok && c
			}
		}
	}
	return false
}

type contextWrapper struct {
	context.Context
	data context.Context
}

func (c contextWrapper) Value(key interface{}) interface{} {
	return c.data.Value(key)
}

func newcontextWrapperWithTimeout(data context.Context, timeout time.Duration) (contextWrapper, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return contextWrapper{
		Context: ctx,
		data:    data,
	}, cancel
}
