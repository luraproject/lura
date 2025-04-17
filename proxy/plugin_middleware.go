// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"fmt"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy/plugin"
)

// NewMwPluginMiddleware returns an endpoint middleware wrapped (if required) with the plugin middleware.
// The plugin middleware will try to load all the required plugins from the register and execute them in order.
func NewMwPluginMiddleware(logger logging.Logger, endpoint *config.EndpointConfig) Middleware {
	cfg, ok := endpoint.ExtraConfig[plugin.MiddlewareNamespace].(map[string]interface{})

	if !ok {
		return emptyMiddlewareFallback(logger)
	}

	return newMwPluginMiddleware(logger, "ENDPOINT", endpoint.Endpoint, cfg)
}

// NewBackendMwPluginMiddleware returns a backend middleware wrapped (if required) with the plugin middleware.
// The plugin middleware will try to load all the required plugins from the register and execute them in order.
func NewBackendMwPluginMiddleware(logger logging.Logger, remote *config.Backend) Middleware {
	cfg, ok := remote.ExtraConfig[plugin.MiddlewareNamespace].(map[string]interface{})

	if !ok {
		return emptyMiddlewareFallback(logger)
	}

	return newMwPluginMiddleware(logger, "BACKEND",
		fmt.Sprintf("%s %s -> %s", remote.ParentEndpointMethod, remote.ParentEndpoint, remote.URLPattern), cfg)
}

func newMwPluginMiddleware(logger logging.Logger, tag, pattern string, cfg map[string]interface{}) Middleware {
	plugins, ok := cfg["name"].([]interface{})
	if !ok {
		return emptyMiddlewareFallback(logger)
	}

	var mws []plugin.MiddlewareFactory

	for _, p := range plugins {
		name, ok := p.(string)
		if !ok {
			continue
		}

		if mf, ok := plugin.GetMiddleware(name); ok {
			mws = append(mws, mf)
		}
	}

	tot := len(mws)
	if tot == 0 {
		return emptyMiddlewareFallback(logger)
	}

	logger.Debug(
		fmt.Sprintf(
			"[%s: %s][Middleware Plugins] Adding %d middlewares",
			tag,
			pattern,
			tot,
		),
	)

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			logger.Fatal("too many proxies for this proxy middleware: newMwPluginMiddleware only accepts 1 proxy, got %d tag: %s, pattern: %s",
				len(next), tag, pattern)
			return nil
		}

		// define the end of the pipe of plugin mw
		p := func(ctx context.Context, r interface{}) (interface{}, error) {
			tmp, ok := r.(RequestWrapper)
			if !ok {
				return nil, fmt.Errorf("unknow type %T", r)
			}

			req := &Request{}

			req.Method = tmp.Method()
			req.URL = tmp.URL()
			req.Query = tmp.Query()
			req.Path = tmp.Path()
			req.Body = tmp.Body()
			req.Params = tmp.Params()
			req.Headers = tmp.Headers()

			resp, err := next[0](ctx, req)

			if resp == nil {
				return nil, err
			}

			return responseWrapper{
				ctx:        ctx,
				request:    r,
				data:       resp.Data,
				isComplete: resp.IsComplete,
				metadata: metadataWrapper{
					headers:    resp.Metadata.Headers,
					statusCode: resp.Metadata.StatusCode,
				},
				io: resp.Io,
			}, err
		}

		// stack all the plugin mws
		for i := tot - 1; i >= 0; i-- {
			p = mws[i](cfg, p)
		}

		// return a wrap over the stacked plugins
		return func(ctx context.Context, r *Request) (*Response, error) {
			rw := newRequestWrapper(ctx, r)
			resp, err := p(ctx, rw)

			if resp == nil {
				return nil, err
			}

			tmp, ok := resp.(ResponseWrapper)
			if !ok {
				return nil, err
			}

			result := &Response{}
			result.Data = tmp.Data()
			result.IsComplete = tmp.IsComplete()
			result.Io = tmp.Io()
			result.Metadata = Metadata{}
			result.Metadata.Headers = tmp.Headers()
			result.Metadata.StatusCode = tmp.StatusCode()
			return result, nil
		}
	}
}
