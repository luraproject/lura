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
	return NewLoadBalancedMiddlewareWithSubscriber(sd.GetRegister().Get(remote.SD)(remote), remote)
}

// NewLoadBalancedMiddlewareWithSubscriber creates proxy middleware adding the most perfomant balancer
// over the received subscriber
func NewLoadBalancedMiddlewareWithSubscriber(subscriber sd.Subscriber, remote *config.Backend) Middleware {
	return newLoadBalancedMiddleware(sd.NewBalancer(subscriber), remote)
}

// NewRoundRobinLoadBalancedMiddleware creates proxy middleware adding a round robin balancer
// over a default subscriber
func NewRoundRobinLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewRoundRobinLoadBalancedMiddlewareWithSubscriber(sd.GetRegister().Get(remote.SD)(remote), remote)
}

// NewRandomLoadBalancedMiddleware creates proxy middleware adding a random balancer
// over a default subscriber
func NewRandomLoadBalancedMiddleware(remote *config.Backend) Middleware {
	return NewRandomLoadBalancedMiddlewareWithSubscriber(sd.GetRegister().Get(remote.SD)(remote), remote)
}

// NewRoundRobinLoadBalancedMiddlewareWithSubscriber creates proxy middleware adding a round robin
// balancer over the received subscriber
func NewRoundRobinLoadBalancedMiddlewareWithSubscriber(subscriber sd.Subscriber, remote *config.Backend) Middleware {
	return newLoadBalancedMiddleware(sd.NewRoundRobinLB(subscriber), remote)
}

// NewRandomLoadBalancedMiddlewareWithSubscriber creates proxy middleware adding a random
// balancer over the received subscriber
func NewRandomLoadBalancedMiddlewareWithSubscriber(subscriber sd.Subscriber, remote *config.Backend) Middleware {
	return newLoadBalancedMiddleware(sd.NewRandomLB(subscriber), remote)
}

func newLoadBalancedMiddleware(lb sd.Balancer, remote *config.Backend) Middleware {
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
				var qp string
				if remote.DisableQueryParametersEncoding {
					q, err := url.QueryUnescape(r.Query.Encode())
					if err != nil {
						return nil, err
					}
					qp = url.PathEscape(q)
				} else {
					qp = r.Query.Encode()
				}
				if len(r.URL.RawQuery) > 0 {
					r.URL.RawQuery += "&" + qp
				} else {
					r.URL.RawQuery += qp
				}
			}

			return next[0](ctx, &r)
		}
	}
}
