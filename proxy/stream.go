package proxy

import (
	"github.com/devopsfaith/krakend/config"
	"net/http"
	"context"
	"github.com/devopsfaith/krakend/encoding"
)

// StreamHTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received HTTPClientFactory
func StreamHTTPProxyFactory(cf HTTPClientFactory) BackendFactory {
	config.ConfigGetters[streamNamespace] = StreamConfigGetter
	return func(backend *config.Backend) Proxy {
		return NewStreamHTTPProxy(backend, cf, backend.Decoder)
	}
}

// NewStreamHTTPProxy creates a http proxy with the injected configuration, HTTPClientFactory and Decoder
func NewStreamHTTPProxy(cfg *config.Backend, clientFactory HTTPClientFactory, decode encoding.Decoder) Proxy {
	return NewHTTPStreamProxyWithHTTPExecutor(cfg, DefaultHTTPRequestExecutor(clientFactory))
}

// NewHTTPStreamProxyWithHTTPExecutor creates a http proxy with the injected configuration, HTTPRequestExecutor and Decoder
func NewHTTPStreamProxyWithHTTPExecutor(cfg *config.Backend, requestExecutor HTTPRequestExecutor) Proxy {
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

		if err != nil {
			return nil, err
		}

		w := NewReadCloserWrapper(ctx, resp.Body)
		metadata := make(map[string]string)

		headers := resp.Header
		for k := range headers {
			metadata[k] = headers.Get(k)
		}

		r := Response{Io: w, IsComplete: true, Metadata: metadata}
		return &r, nil
	}
}
