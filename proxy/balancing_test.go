package proxy

import (
	"context"
	"errors"
	"testing"
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

type dummyBalancer string

func (d dummyBalancer) Host() (string, error) { return string(d), nil }

type explosiveBalancer struct {
	Error error
}

func (e explosiveBalancer) Host() (string, error) { return "", e.Error }
