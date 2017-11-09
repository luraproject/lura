package proxy

import (
	"bytes"
	"context"
	"fmt"
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

func TestWrapper(t *testing.T) {
	expected := "supu"
	closed := false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := NewReadCloserWrapper(ctx, dummyRC{bytes.NewBufferString(expected), &closed})
	var out bytes.Buffer
	tot, err := out.ReadFrom(r)
	if err != nil {
		t.Errorf("Total bits read: %d. Err: %s", tot, err.Error())
		return
	}
	if closed {
		t.Error("The subject shouldn't be closed yet")
		return
	}
	if tot != 4 {
		t.Errorf("Unexpected number of bits read: %d", tot)
		return
	}
	if v := out.String(); v != expected {
		t.Errorf("Unexpected content: %s", v)
		return
	}

	cancel()
	<-time.After(100 * time.Millisecond)
	if !closed {
		t.Error("The subject should be already closed")
		return
	}
}

type dummyRC struct {
	r      io.Reader
	closed *bool
}

func (d dummyRC) Read(b []byte) (int, error) {
	if *(d.closed) {
		return -1, fmt.Errorf("Reading from a closed source")
	}
	return d.r.Read(b)
}

func (d dummyRC) Close() error {
	*(d.closed) = true
	return nil
}
