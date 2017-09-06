package streaming

import (
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"net/http"
	"context"
)

var streamHttpProxy = CustomStreamHTTPProxyFactory(proxy.NewHTTPClient)

// CustomStreamHTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received HTTPClientFactory
func CustomStreamHTTPProxyFactory(cf proxy.HTTPClientFactory) proxy.BackendFactory {
	return func(backend *config.Backend) proxy.Proxy {
		return NewStreamHTTPProxy(backend, cf)
	}
}

// StreamHTTPProxyFactory returns a BackendFactory. The Proxies it creates will use the received HTTPClientFactory
func StreamHTTPProxyFactory(client *http.Client) proxy.BackendFactory {
	return func(backend *config.Backend) proxy.Proxy {
		return NewStreamHTTPProxy(backend, func(_ context.Context) *http.Client { return client })
	}
}

// NewStreamHTTPProxy creates a streaming http proxy with the injected configuration, HTTPClientFactory and Decoder
func NewStreamHTTPProxy(cfg *config.Backend, clientFactory proxy.HTTPClientFactory) proxy.Proxy {
	return NewHTTPStreamProxyWithHTTPExecutor(cfg, proxy.DefaultHTTPRequestExecutor(clientFactory))
}

// NewHTTPStreamProxyWithHTTPExecutor creates a streaming http proxy with the injected configuration, HTTPRequestExecutor and Decoder
func NewHTTPStreamProxyWithHTTPExecutor(cfg *config.Backend, requestExecutor proxy.HTTPRequestExecutor) proxy.Proxy {
	return func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
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
			return nil, proxy.ErrInvalidStatusCode
		}

		if err != nil {
			return nil, err
		}

		w := proxy.NewReadCloserWrapper(ctx, resp.Body)
		metadata := make(map[string]string)

		headers := resp.Header
		for k := range headers {
			metadata[k] = headers.Get(k)
		}

		r := proxy.Response{Io: w, IsComplete: true, Metadata: metadata}
		return &r, nil
	}
}
