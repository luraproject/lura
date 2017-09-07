package streaming

import (
	"context"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"net/http"
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
	return proxy.NewHTTPProxyDetailed(cfg, proxy.DefaultHTTPRequestExecutor(clientFactory), StreamHTTPResponseParser())
}

type streamHTTPResponseParser struct {
}

func StreamHTTPResponseParser() proxy.HTTPResponseParser {
	return streamHTTPResponseParser{}
}

func (p streamHTTPResponseParser) HandleResponse(ctx context.Context, resp *http.Response) (*proxy.Response, error) {
	w := proxy.NewReadCloserWrapper(ctx, resp.Body)
	metadata := make(map[string]string)

	headers := resp.Header
	for k := range headers {
		metadata[k] = headers.Get(k)
	}

	return &proxy.Response{Io: w, IsComplete: true, Metadata: metadata}, nil

}
