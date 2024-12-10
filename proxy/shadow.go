// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

const (
	shadowKey        = "shadow"
	shadowTimeoutKey = "shadow_timeout"
)

type shadowFactory struct {
	f Factory
}

// New check the Backends for an ExtraConfig with the "shadow" param to true
// implements the Factory interface. Sets the "shadow_timeout" defined in the
// config; uses the backend timeout as fallback.
func (s shadowFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	if len(cfg.Backend) == 0 {
		err = ErrNoBackends
		return
	}

	cfgCopy := *cfg

	var shadow []*config.Backend
	var regular []*config.Backend
	var maxTimeout time.Duration
	for _, b := range cfgCopy.Backend {
		if d, ok := isShadowBackend(b); ok {
			if maxTimeout < d {
				maxTimeout = d
			}
			shadow = append(shadow, b)
			continue
		}
		regular = append(regular, b)
	}

	cfgCopy.Backend = regular
	p, err = s.f.New(&cfgCopy)

	if len(shadow) > 0 {
		cfgCopy.Backend = shadow
		pShadow, _ := s.f.New(&cfgCopy)
		p = ShadowMiddlewareWithTimeout(maxTimeout, p, pShadow)
	}

	return
}

// NewShadowFactory creates a new shadowFactory using the provided Factory
func NewShadowFactory(f Factory) Factory {
	return shadowFactory{f}
}

// ShadowMiddlewareWithLogger is a Middleware that creates a shadowProxy
func ShadowMiddlewareWithLogger(logger logging.Logger, next ...Proxy) Proxy {
	switch len(next) {
	case 0:
		logger.Fatal("not enough proxies for this endpoint: ShadowMiddlewareWithLogger only accepts 1 or 2 proxies, got 0")
		return nil
	case 1:
		return next[0]
	case 2:
		return NewShadowProxy(next[0], next[1])
	default:
		logger.Fatal("too many proxies for this proxy middleware: ShadowMiddlewareWithLogger only accepts 1 or 2 proxies, got %d", len(next))
		return nil
	}
}

// ShadowMiddleware is a Middleware that creates a shadowProxy
func ShadowMiddleware(next ...Proxy) Proxy {
	return ShadowMiddlewareWithLogger(logging.NoOp, next...)
}

// ShadowMiddlewareWithTimeoutAndLogger is a Middleware that creates a shadowProxy with a timeout in the context
func ShadowMiddlewareWithTimeoutAndLogger(logger logging.Logger, timeout time.Duration, next ...Proxy) Proxy {
	switch len(next) {
	case 0:
		logger.Fatal("not enough proxies for this endpoint: ShadowMiddlewareWithTimeoutAndLogger only accepts 1 or 2 proxies, got 0")
		return nil
	case 1:
		return next[0]
	case 2:
		return NewShadowProxyWithTimeout(timeout, next[0], next[1])
	default:
		logger.Fatal("too many proxies for this proxy middleware: ShadowMiddlewareWithTimeoutAndLogger only accepts 1 or 2 proxies, got %d", len(next))
		return nil
	}
}

// ShadowMiddlewareWithTimeout is a Middleware that creates a shadowProxy with a timeout in the context
func ShadowMiddlewareWithTimeout(timeout time.Duration, next ...Proxy) Proxy {
	return ShadowMiddlewareWithTimeoutAndLogger(logging.NoOp, timeout, next...)
}

// NewShadowProxy returns a Proxy that sends requests to p1 and p2 but ignores
// the response of p2.
func NewShadowProxy(p1, p2 Proxy) Proxy {
	return NewShadowProxyWithTimeout(config.DefaultTimeout, p1, p2)
}

// NewShadowProxyWithTimeout returns a Proxy that sends requests to p1 and p2 but ignores
// the response of p2. Sets a timeout in the context.
func NewShadowProxyWithTimeout(timeout time.Duration, p1, p2 Proxy) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		shadowCtx, cancel := newContextWrapperWithTimeout(ctx, timeout)
		shadowRequest := CloneRequest(request)
		go func() {
			p2(shadowCtx, shadowRequest)
			cancel()
		}()
		return p1(ctx, request)
	}
}

func isShadowBackend(c *config.Backend) (time.Duration, bool) {
	duration := c.Timeout
	v, ok := c.ExtraConfig[Namespace]
	if !ok {
		return duration, false
	}

	e, ok := v.(map[string]interface{})
	if !ok {
		return duration, false
	}

	k, ok := e[shadowKey]
	if !ok {
		return duration, false
	}

	if s, ok := k.(bool); !ok || !s {
		return duration, false
	}

	t, ok := e[shadowTimeoutKey].(string)
	if !ok {
		return duration, true
	}

	if d, err := time.ParseDuration(t); err == nil {
		duration = d
	}

	return duration, true
}

type contextWrapper struct {
	context.Context
	data context.Context
}

func (c contextWrapper) Value(key interface{}) interface{} {
	return c.data.Value(key)
}

func newContextWrapperWithTimeout(data context.Context, timeout time.Duration) (contextWrapper, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return contextWrapper{
		Context: ctx,
		data:    data,
	}, cancel
}
