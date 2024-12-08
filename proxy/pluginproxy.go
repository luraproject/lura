package proxy

import (
	"context"
	"fmt"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy/pluginproxy"
)

func NewPluginProxyMiddleware(logger logging.Logger, endpoint *config.EndpointConfig) Middleware {
	cfg, ok := endpoint.ExtraConfig[pluginproxy.Namespace].(map[string]interface{})
	if !ok {
		return emptyMiddlewareFallback(logger)
	}
	return newPluginProxyMiddleware(logger, "ENDPOINT", endpoint.Endpoint, cfg)
}

func NewBackendPluginProxyMiddleware(logger logging.Logger, remote *config.Backend) Middleware {
	cfg, ok := remote.ExtraConfig[pluginproxy.Namespace].(map[string]interface{})
	if !ok {
		return emptyMiddlewareFallback(logger)
	}
	return newPluginProxyMiddleware(logger, "BACKEND",
		fmt.Sprintf("%s %s -> %s", remote.ParentEndpointMethod, remote.ParentEndpoint, remote.URLPattern), cfg)
}

func newPluginProxyMiddleware(logger logging.Logger, tag, pattern string, cfg map[string]interface{}) Middleware {
	plugins, ok := cfg["name"].([]interface{})
	if !ok {
		return emptyMiddlewareFallback(logger)
	}

	var proxies []pluginproxy.Handler

	for _, p := range plugins {
		name, ok := p.(string)
		if !ok {
			continue
		}

		h, ok := pluginproxy.GetProxy(name)
		if !ok {
			continue
		}
		proxies = append(proxies, h)
	}

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			logger.Fatal("too many proxies for this proxy middleware: newPluginProxyMiddleware only accepts 1 proxy, got %d tag: %s, pattern: %s",
				len(next), tag, pattern)
			return nil
		}

		return func(ctx context.Context, r *Request) (*Response, error) {
			proxies = append(proxies, func(context.Context, map[string]interface{}, pluginproxy.ProxyWrapper) pluginproxy.ProxyWrapper {
				return func(ctx context.Context, rw pluginproxy.RequestWrapper) (pluginproxy.ResponseWrapper, error) {
					r := Request{
						Method:  rw.Method(),
						URL:     rw.URL(),
						Query:   rw.Query(),
						Path:    rw.Path(),
						Body:    rw.Body(),
						Params:  rw.Params(),
						Headers: rw.Headers(),
					}
					resp, err := next[0](ctx, &r)
					return newResponseWrapper(ctx, resp), err
				}
			})
			return executeProxies(ctx, r, cfg, proxies)
		}
	}
}

// executeProxies executes all proxies and expecting the last proxies to be not calling again the next proxy.
func executeProxies(ctx context.Context, r *Request, cfg map[string]interface{}, proxies []pluginproxy.Handler) (*Response, error) {
	var proxy pluginproxy.ProxyWrapper

	for i := len(proxies) - 1; i >= 0; i-- {
		proxy = proxies[i](ctx, cfg, proxy)
	}

	resp, err := proxy(ctx, newRequestWrapper(ctx, r))
	if err != nil {
		return nil, err
	}

	return &Response{
		Data:       resp.Data(),
		IsComplete: resp.IsComplete(),
		Io:         resp.Io(),
		Metadata: Metadata{
			Headers:    resp.Headers(),
			StatusCode: resp.StatusCode(),
		},
	}, nil
}
