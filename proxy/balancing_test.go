// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/sd/dnssrv"
)

func TestNewLoadBalancedMiddleware_ok(t *testing.T) {
	want := "supu:8080/tupu"
	lb := newLoadBalancedMiddleware(dummyBalancer("supu:8080"))
	assertion := func(ctx context.Context, request *Request) (*Response, error) {
		if request.URL.String() != want {
			t.Errorf("The middleware did not update the request URL! want [%s], have [%s]\n", want, request.URL)
		}
		return nil, nil
	}
	if _, err := lb(assertion)(context.Background(), &Request{
		Path: "/tupu",
	}); err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
}

func TestNewLoadBalancedMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	lb := newLoadBalancedMiddleware(dummyBalancer("supu"))
	lb(explosiveProxy(t), explosiveProxy(t))
}

func TestNewLoadBalancedMiddleware_explosiveBalancer(t *testing.T) {
	expected := errors.New("supu")
	lb := newLoadBalancedMiddleware(explosiveBalancer{expected})
	if _, err := lb(explosiveProxy(t))(context.Background(), &Request{}); err != expected {
		t.Errorf("The middleware did not propagate the lb error\n")
	}
}

func TestNewRoundRobinLoadBalancedMiddleware(t *testing.T) {
	testLoadBalancedMw(t, NewRoundRobinLoadBalancedMiddleware(&config.Backend{
		Host: []string{"http://127.0.0.1:8080"},
	}))
}

func TestNewRandomLoadBalancedMiddleware(t *testing.T) {
	testLoadBalancedMw(t, NewRandomLoadBalancedMiddleware(&config.Backend{
		Host: []string{"http://127.0.0.1:8080"},
	}))
}

func testLoadBalancedMw(t *testing.T, lb Middleware) {
	for _, tc := range []struct {
		path     string
		query    url.Values
		expected string
	}{
		{
			path:     "/tupu",
			expected: "http://127.0.0.1:8080/tupu",
		},
		{
			path:     "/tupu?extra=true",
			expected: "http://127.0.0.1:8080/tupu?extra=true",
		},
		{
			path:     "/tupu?extra=true",
			query:    url.Values{"some": []string{"none"}},
			expected: "http://127.0.0.1:8080/tupu?extra=true&some=none",
		},
		{
			path:     "/tupu",
			query:    url.Values{"some": []string{"none"}},
			expected: "http://127.0.0.1:8080/tupu?some=none",
		},
	} {
		assertion := func(ctx context.Context, request *Request) (*Response, error) {
			if request.URL.String() != tc.expected {
				t.Errorf("The middleware did not update the request URL! want [%s], have [%s]\n", tc.expected, request.URL)
			}
			return nil, nil
		}
		if _, err := lb(assertion)(context.Background(), &Request{
			Path:  tc.path,
			Query: tc.query,
		}); err != nil {
			t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
		}
	}
}

func TestNewLoadBalancedMiddleware_parsingError(t *testing.T) {
	lb := NewRandomLoadBalancedMiddleware(&config.Backend{
		Host: []string{"127.0.0.1:8080"},
	})
	assertion := func(ctx context.Context, request *Request) (*Response, error) {
		t.Error("The middleware didn't block the request!")
		return nil, nil
	}
	if _, err := lb(assertion)(context.Background(), &Request{
		Path: "/tupu",
	}); err == nil {
		t.Error("The middleware didn't propagate the expected error")
	}
}

func TestNewRoundRobinLoadBalancedMiddleware_DNSSRV(t *testing.T) {
	defaultLookup := dnssrv.DefaultLookup

	dnssrv.DefaultLookup = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", []*net.SRV{
			{
				Port:   8080,
				Target: "127.0.0.1",
				Weight: 1,
			},
		}, nil
	}
	testLoadBalancedMw(t, NewRoundRobinLoadBalancedMiddlewareWithSubscriber(dnssrv.New("some.service.example.tld")))

	dnssrv.DefaultLookup = defaultLookup
}

type dummyBalancer string

func (d dummyBalancer) Host() (string, error) { return string(d), nil }

type explosiveBalancer struct {
	Error error
}

func (e explosiveBalancer) Host() (string, error) { return "", e.Error }
