// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"net/url"
	"strings"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/sd"
)

// NewLoadBalancedMiddleware creates proxy middleware adding the most perfomant balancer
// over a default subscriber
func NewLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewLoadBalancedMiddlewareWithSubscriber(sd.GetRegister().Get(remote.SD)(remote))
}

// NewLoadBalancedMiddlewareWithSubscriber creates proxy middleware adding the most perfomant balancer
// over the received subscriber
func NewLoadBalancedMiddlewareWithSubscriber(subscriber sd.Subscriber) Middleware {
	return newLoadBalancedMiddleware(sd.NewBalancer(subscriber))
}

// NewRoundRobinLoadBalancedMiddleware creates proxy middleware adding a round robin balancer
// over a default subscriber
func NewRoundRobinLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewRoundRobinLoadBalancedMiddlewareWithSubscriber(sd.GetRegister().Get(remote.SD)(remote))
}

// NewRandomLoadBalancedMiddleware creates proxy middleware adding a random balancer
// over a default subscriber
func NewRandomLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewRandomLoadBalancedMiddlewareWithSubscriber(sd.GetRegister().Get(remote.SD)(remote))
}

// NewRoundRobinLoadBalancedMiddlewareWithSubscriber creates proxy middleware adding a round robin
// balancer over the received subscriber
func NewRoundRobinLoadBalancedMiddlewareWithSubscriber(subscriber sd.Subscriber) Middleware {
	return newLoadBalancedMiddleware(sd.NewRoundRobinLB(subscriber))
}

// NewRandomLoadBalancedMiddlewareWithSubscriber creates proxy middleware adding a random
// balancer over the received subscriber
func NewRandomLoadBalancedMiddlewareWithSubscriber(subscriber sd.Subscriber) Middleware {
	return newLoadBalancedMiddleware(sd.NewRandomLB(subscriber))
}

func newLoadBalancedMiddleware(lb sd.Balancer) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			host, err := lb.Host()
			if err != nil {
				return nil, err
			}
			r := request.Clone()

			var b strings.Builder
			b.WriteString(host)
			b.WriteString(r.Path)
			r.URL, err = url.Parse(b.String())
			if err != nil {
				return nil, err
			}
			if len(r.Query) > 0 {
				if len(r.URL.RawQuery) > 0 {
					r.URL.RawQuery += "&" + r.Query.Encode()
				} else {
					r.URL.RawQuery += r.Query.Encode()
				}
			}

			return next[0](ctx, &r)
		}
	}
}
