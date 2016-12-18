package proxy

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestEmptyMiddleware_ok(t *testing.T) {
	expected := Response{}
	result, err := EmptyMiddleware(dummyProxy(&expected))(context.Background(), &Request{})
	if err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if result != &expected {
		t.Errorf("The middleware returned an unexpected result: %v\n", result)
	}
}

func TestEmptyMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	EmptyMiddleware(NoopProxy, NoopProxy)
}

func explosiveProxy(t *testing.T) Proxy {
	return func(ctx context.Context, _ *Request) (*Response, error) {
		t.Error("This proxy shouldn't been executed!")
		return &Response{}, nil
	}
}

func dummyProxy(r *Response) Proxy {
	return func(_ context.Context, _ *Request) (*Response, error) {
		return r, nil
	}
}

func delayedProxy(t *testing.T, timeout time.Duration, r *Response) Proxy {
	return func(ctx context.Context, _ *Request) (*Response, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(timeout):
			return r, nil
		}
	}
}

func newDummyReadCloser(content string) io.ReadCloser {
	return dummyReadCloser{strings.NewReader(content)}
}

type dummyReadCloser struct {
	reader io.Reader
}

func (d dummyReadCloser) Read(p []byte) (int, error) {
	return d.reader.Read(p)
}

func (d dummyReadCloser) Close() error {
	return nil
}
