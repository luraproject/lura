package httpcache

import (
	"context"
	"net/http"

	"github.com/gregjones/httpcache"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
)

var client http.Client

// NewHTTPProxy implements the proxy.BackendFactory. The returned proxy.Proxy comes with an injected
// NewHTTPClient using a custom transport layer provided by the httpcache package
func NewHTTPProxy(tp *httpcache.Transport) proxy.BackendFactory {
	client = http.Client{Transport: tp}
	return func(backend *config.Backend) proxy.Proxy {
		return proxy.NewHTTPProxy(backend, NewHTTPClient, backend.Decoder)
	}
}

// NewHTTPClient implements the proxy.HTTPClientFactory interface
func NewHTTPClient(_ context.Context) *http.Client { return &client }
