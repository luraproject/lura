package proxy

import (
	"context"
	"errors"
	"net/http"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
)

// ErrInvalidStatusCode is the error returned by the http proxy when the received status code
// is not a 200 nor a 201
var ErrInvalidStatusCode = errors.New("Invalid status code")

var httpProxy = CustomHTTPProxyFactory(NewHTTPClient)

// HTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received net/http.Client
func HTTPProxyFactory(client *http.Client) BackendFactory {
	return CustomHTTPProxyFactory(func(_ context.Context) *http.Client { return client })
}

// CustomHTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received HTTPClientFactory
func CustomHTTPProxyFactory(cf HTTPClientFactory) BackendFactory {
	return func(backend *config.Backend) Proxy {
		return NewHTTPProxy(backend, cf, backend.Decoder)
	}
}

// NewHTTPProxy creates a http proxy with the injected configuration, HTTPClientFactory and Decoder
func NewHTTPProxy(remote *config.Backend, clientFactory HTTPClientFactory, decode encoding.Decoder) Proxy {
	return NewHTTPProxyWithHTTPExecutor(remote, DefaultHTTPRequestExecutor(clientFactory), decode)
}

// NewHTTPProxyWithHTTPExecutor creates a http proxy with the injected configuration, HTTPRequestExecutor and Decoder
func NewHTTPProxyWithHTTPExecutor(remote *config.Backend, requestExecutor HTTPRequestExecutor, dec encoding.Decoder) Proxy {
	if remote.Encoding == encoding.NOOP {
		return NewHTTPProxyDetailed(remote, requestExecutor, NoOpHTTPStatusHandler, NoOpHTTPResponseParser)
	}

	ef := NewEntityFormatter(remote)
	rp := DefaultHTTPResponseParserFactory(HTTPResponseParserConfig{dec, ef})
	return NewHTTPProxyDetailed(remote, requestExecutor, DefaultHTTPStatusHandler, rp)
}

// NewHTTPProxyDetailed creates a http proxy with the injected configuration, HTTPRequestExecutor, Decoder and HTTPResponseParser
func NewHTTPProxyDetailed(remote *config.Backend, requestExecutor HTTPRequestExecutor, ch HTTPStatusHandler, rp HTTPResponseParser) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		requestToBakend, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
		if err != nil {
			return nil, err
		}
		requestToBakend.Header = make(map[string][]string, len(request.Headers))
		for k, vs := range request.Headers {
			tmp := make([]string, len(vs))
			copy(tmp, vs)
			requestToBakend.Header[k] = tmp
		}

		resp, err := requestExecutor(ctx, requestToBakend)
		requestToBakend.Body.Close()
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
			return nil, err
		}

		return rp(ctx, resp)
	}
}

// NewRequestBuilderMiddleware creates a proxy middleware that parses the request params received
// from the outter layer and generates the path to the backend endpoints
func NewRequestBuilderMiddleware(remote *config.Backend) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			r := request.Clone()
			r.GeneratePath(remote.URLPattern)
			r.Method = remote.Method
			return next[0](ctx, &r)
		}
	}
}
