package proxy

import (
	"context"
	"errors"
	"net/http"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
)

var ErrInvalidStatusCode = errors.New("Invalid status code")

// HTTPClientFactory creates http clients based with the received context
type HTTPClientFactory func(ctx context.Context) *http.Client

// NewHTTPClient just creates a http default client
func NewHTTPClient(_ context.Context) *http.Client { return http.DefaultClient }

func httpProxy(backend *config.Backend) Proxy {
	return NewHTTPProxy(backend, NewHTTPClient, backend.Decoder)
}

// NewRequestBuilderMiddleware creates a proxy middleware that parses the request params recived
// the user and generates the path to the backend endpoints
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

// NewHTTPProxy creates a http proxy with the injected configuration, HTTPClientFactory and Decoder
func NewHTTPProxy(remote *config.Backend, clientFactory HTTPClientFactory, decode encoding.Decoder) Proxy {
	formatter := NewEntityFormatter(remote.Target, remote.Whitelist, remote.Blacklist, remote.Group, remote.Mapping)

	return func(ctx context.Context, request *Request) (*Response, error) {
		requestToBakend, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
		if err != nil {
			return nil, err
		}
		requestToBakend.Header = request.Headers

		resp, err := clientFactory(ctx).Do(requestToBakend.WithContext(ctx))
		requestToBakend.Body.Close()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return nil, ErrInvalidStatusCode
		}

		var data map[string]interface{}
		err = decode(resp.Body, &data)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		r := formatter.Format(Response{data, true})
		return &r, nil
	}
}
