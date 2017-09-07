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

// HTTPClientFactory creates http clients based with the received context
type HTTPClientFactory func(ctx context.Context) *http.Client

// NewHTTPClient just creates a http default client
func NewHTTPClient(_ context.Context) *http.Client { return http.DefaultClient }

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

// HTTPRequestExecutor defines the interface of the request executor for the HTTP transport protocol
type HTTPRequestExecutor func(ctx context.Context, req *http.Request) (*http.Response, error)

// DefaultHTTPRequestExecutor creates a HTTPRequestExecutor with the received HTTPClientFactory
func DefaultHTTPRequestExecutor(clientFactory HTTPClientFactory) HTTPRequestExecutor {
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return clientFactory(ctx).Do(req.WithContext(ctx))
	}
}

// NewHTTPProxy creates a http proxy with the injected configuration, HTTPClientFactory and Decoder
func NewHTTPProxy(remote *config.Backend, clientFactory HTTPClientFactory, decode encoding.Decoder) Proxy {
	return NewHTTPProxyWithHTTPExecutor(remote, DefaultHTTPRequestExecutor(clientFactory), decode)
}

// NewHTTPProxyWithHTTPExecutor creates a http proxy with the injected configuration, HTTPRequestExecutor and Decoder
func NewHTTPProxyWithHTTPExecutor(remote *config.Backend, requestExecutor HTTPRequestExecutor, decode encoding.Decoder) Proxy {
	formatter := NewEntityFormatter(remote.Target, remote.Whitelist, remote.Blacklist, remote.Group, remote.Mapping)
	return NewHTTPProxyDetailed(remote, requestExecutor, DefaultHTTPResponseParser(decode, formatter))
}

// HTTPResponseParser defines the interface of the response parser for the HTTP transport protocol
type HTTPResponseParser interface {
	HandleResponse(context.Context, *http.Response) (*Response, error)
}

// NewHTTPProxyDetailed creates a http proxy with the injected configuration, HTTPRequestExecutor, Decoder and HTTPResponseParser
func NewHTTPProxyDetailed(remote *config.Backend, requestExecutor HTTPRequestExecutor, responseParser HTTPResponseParser) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		requestToBakend, err := http.NewRequest(request.Method, request.URL.String(), request.Body)
		if err != nil {
			return nil, err
		}
		requestToBakend.Header = request.Headers

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
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return nil, ErrInvalidStatusCode
		}

		return responseParser.HandleResponse(ctx, resp)
	}
}

type defaultHTTPResponseParser struct {
	decoder   encoding.Decoder
	formatter EntityFormatter
}

func DefaultHTTPResponseParser(decoder encoding.Decoder, formatter EntityFormatter) HTTPResponseParser {
	return defaultHTTPResponseParser{decoder, formatter}
}

func (p defaultHTTPResponseParser) HandleResponse(ctx context.Context, resp *http.Response) (*Response, error) {
	var data map[string]interface{}
	err := p.decoder(resp.Body, &data)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	newResponse := Response{Data: data, IsComplete: true}
	newResponse = p.formatter.Format(newResponse)
	return &newResponse, nil

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
