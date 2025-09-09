// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/encoding"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/transport/http/client"
)

var httpProxy = CustomHTTPProxyFactory(client.NewHTTPClient)

// HTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received net/http.Client
func HTTPProxyFactory(client *http.Client) BackendFactory {
	return CustomHTTPProxyFactory(func(_ context.Context) *http.Client { return client })
}

// CustomHTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received HTTPClientFactory
func CustomHTTPProxyFactory(cf client.HTTPClientFactory) BackendFactory {
	return func(backend *config.Backend) Proxy {
		return NewHTTPProxy(backend, cf, backend.Decoder)
	}
}

// NewHTTPProxy creates a http proxy with the injected configuration, HTTPClientFactory and Decoder
func NewHTTPProxy(remote *config.Backend, cf client.HTTPClientFactory, decode encoding.Decoder) Proxy {
	return NewHTTPProxyWithHTTPExecutor(remote, client.DefaultHTTPRequestExecutor(cf), decode)
}

// NewHTTPProxyWithHTTPExecutor creates a http proxy with the injected configuration, HTTPRequestExecutor and Decoder
func NewHTTPProxyWithHTTPExecutor(remote *config.Backend, re client.HTTPRequestExecutor, dec encoding.Decoder) Proxy {
	if remote.Encoding == encoding.NOOP {
		return NewHTTPProxyDetailed(remote, re, client.NoOpHTTPStatusHandler, NoOpHTTPResponseParser)
	}

	ef := NewEntityFormatter(remote)
	rp := DefaultHTTPResponseParserFactory(HTTPResponseParserConfig{dec, ef})
	return NewHTTPProxyDetailed(remote, re, client.GetHTTPStatusHandler(remote), rp)
}

const (
	clientHTTPOptions            string = "backend/http/client"
	clientHTTPOptionRedirectPost string = "send_body_on_redirect"
)

// redirectPostReaderFactory checks if the clientHTTPOptionRedirectPost is enabled
// This will read the body and return a bytes.Buffer with the body content, so we
// delegate to http.NewRequest the population of request.GetBody so a redirect (307
// and 308) is executed while maintaining the method and the body
// This is necessary since the request comes from another http.Client and it's not
// a concrete type that can be copied but just a io.ReaderCloser (*http.body)
func redirectPostReaderFactory(cfg *config.Backend) func(r io.ReadCloser) io.Reader {
	emptyFactory := func(r io.ReadCloser) io.Reader { return r }
	if cfg == nil || cfg.ExtraConfig == nil {
		return emptyFactory
	}
	v, ok := cfg.ExtraConfig[clientHTTPOptions].(map[string]interface{})
	if !ok {
		return emptyFactory
	}
	if opt, ok := v[clientHTTPOptionRedirectPost].(bool); !ok || !opt {
		return emptyFactory
	}
	return func(r io.ReadCloser) io.Reader {
		if r == http.NoBody || r == nil {
			return r
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(r)
		r.Close()
		return buf
	}
}

// NewHTTPProxyDetailed creates a http proxy with the injected configuration, HTTPRequestExecutor,
// Decoder and HTTPResponseParser
func NewHTTPProxyDetailed(cfg *config.Backend, re client.HTTPRequestExecutor, ch client.HTTPStatusHandler, rp HTTPResponseParser) Proxy {
	bodyFactory := redirectPostReaderFactory(cfg)
	return func(ctx context.Context, request *Request) (*Response, error) {
		requestToBackend, err := http.NewRequest(strings.ToTitle(request.Method), request.URL.String(), bodyFactory(request.Body))
		if err != nil {
			return nil, err
		}
		requestToBackend.Header = make(map[string][]string, len(request.Headers))
		for k, vs := range request.Headers {
			tmp := make([]string, len(vs))
			copy(tmp, vs)
			requestToBackend.Header[k] = tmp
		}
		if request.Body != nil {
			if v, ok := request.Headers["Content-Length"]; ok && len(v) == 1 && v[0] != "chunked" {
				if size, err := strconv.Atoi(v[0]); err == nil {
					requestToBackend.ContentLength = int64(size)
				}
			}
		}

		resp, err := re(ctx, requestToBackend)
		if requestToBackend.Body != nil {
			requestToBackend.Body.Close()
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if err != nil {
			return nil, err
		}

		resp, err = ch(ctx, resp)
		if err != nil {
			if t, ok := err.(responseError); ok {
				return &Response{
					Data: map[string]interface{}{
						fmt.Sprintf("error_%s", t.Name()): t,
					},
					Metadata: Metadata{StatusCode: t.StatusCode()},
				}, nil
			}
			return nil, err
		}

		return rp(ctx, resp)
	}
}

// NewRequestBuilderMiddleware creates a proxy middleware that parses the request params received
// from the outer layer and generates the path to the backend endpoints
var NewRequestBuilderMiddleware = func(remote *config.Backend) Middleware {
	return newRequestBuilderMiddleware(logging.NoOp, remote)
}

func NewRequestBuilderMiddlewareWithLogger(logger logging.Logger, remote *config.Backend) Middleware {
	return newRequestBuilderMiddleware(logger, remote)
}

func newRequestBuilderMiddleware(l logging.Logger, remote *config.Backend) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			l.Fatal("too many proxies for this %s %s -> %s proxy middleware: newRequestBuilderMiddleware only accepts 1 proxy, got %d", remote.ParentEndpointMethod, remote.ParentEndpoint, remote.URLPattern, len(next))
			return nil
		}
		return func(ctx context.Context, r *Request) (*Response, error) {
			r.GeneratePath(remote.URLPattern)
			r.Method = remote.Method
			return next[0](ctx, r)
		}
	}
}

type responseError interface {
	Error() string
	Name() string
	StatusCode() int
}
