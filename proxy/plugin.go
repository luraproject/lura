// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy/plugin"
)

// NewPluginMiddleware returns an endpoint middleware wrapped (if required) with the plugin middleware.
// The plugin middleware will try to load all the required plugins from the register and execute them in order.
// RequestModifiers are executed before passing the request to the next middlware. ResponseModifiers are executed
// once the response is returned from the next middleware.
func NewPluginMiddleware(logger logging.Logger, endpoint *config.EndpointConfig) Middleware {
	cfg, ok := endpoint.ExtraConfig[plugin.Namespace].(map[string]interface{})

	if !ok {
		return emptyMiddlewareFallback(logger)
	}

	return newPluginMiddleware(logger, "ENDPOINT", endpoint.Endpoint, cfg)
}

// NewBackendPluginMiddleware returns a backend middleware wrapped (if required) with the plugin middleware.
// The plugin middleware will try to load all the required plugins from the register and execute them in order.
// RequestModifiers are executed before passing the request to the next middlware. ResponseModifiers are executed
// once the response is returned from the next middleware.
func NewBackendPluginMiddleware(logger logging.Logger, remote *config.Backend) Middleware {
	cfg, ok := remote.ExtraConfig[plugin.Namespace].(map[string]interface{})

	if !ok {
		return emptyMiddlewareFallback(logger)
	}

	return newPluginMiddleware(logger, "BACKEND",
		fmt.Sprintf("%s %s -> %s", remote.ParentEndpointMethod, remote.ParentEndpoint, remote.URLPattern), cfg)
}

func newPluginMiddleware(logger logging.Logger, tag, pattern string, cfg map[string]interface{}) Middleware {
	plugins, ok := cfg["name"].([]interface{})
	if !ok {
		return emptyMiddlewareFallback(logger)
	}

	var reqModifiers []func(interface{}) (interface{}, error)

	var respModifiers []func(interface{}) (interface{}, error)

	for _, p := range plugins {
		name, ok := p.(string)
		if !ok {
			continue
		}

		if mf, ok := plugin.GetRequestModifier(name); ok {
			if fn := mf(cfg); fn != nil {
				reqModifiers = append(reqModifiers, fn)
			}
			continue
		}

		if mf, ok := plugin.GetResponseModifier(name); ok {
			if fn := mf(cfg); fn != nil {
				respModifiers = append(respModifiers, fn)
			}
		}
	}

	totReqModifiers, totRespModifiers := len(reqModifiers), len(respModifiers)
	if totReqModifiers == totRespModifiers && totRespModifiers == 0 {
		return emptyMiddlewareFallback(logger)
	}

	logger.Debug(
		fmt.Sprintf(
			"[%s: %s][Modifier Plugins] Adding %d request and %d response modifiers",
			tag,
			pattern,
			totReqModifiers,
			totRespModifiers,
		),
	)

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			logger.Fatal("too many proxies for this proxy middleware: newPluginMiddleware only accepts 1 proxy, got %d tag: %s, pattern: %s",
				len(next), tag, pattern)
			return nil
		}

		if totReqModifiers == 0 {
			return func(ctx context.Context, r *Request) (*Response, error) {
				resp, err := next[0](ctx, r)
				if err != nil {
					return resp, err
				}

				return executeResponseModifiers(ctx, respModifiers, resp, newRequestWrapper(ctx, r))
			}
		}

		if totRespModifiers == 0 {
			return func(ctx context.Context, r *Request) (*Response, error) {
				req, resp, err := executeRequestModifiers(ctx, reqModifiers, r)
				if err != nil {
					return nil, err
				}

				if resp != nil {
					return resp, nil
				}

				return next[0](ctx, req)
			}
		}

		return func(ctx context.Context, r *Request) (*Response, error) {
			req, resp, err := executeRequestModifiers(ctx, reqModifiers, r)
			if err != nil {
				return nil, err
			}

			if resp == nil {
				var err error
				resp, err = next[0](ctx, req)
				if err != nil {
					return resp, err
				}
			}

			return executeResponseModifiers(ctx, respModifiers, resp, newRequestWrapper(ctx, req))
		}
	}
}

func executeRequestModifiers(ctx context.Context, reqModifiers []func(interface{}) (interface{}, error), req *Request) (*Request, *Response, error) {
	var tmp RequestWrapper
	tmp = newRequestWrapper(ctx, req)
	var resp *Response

	for _, f := range reqModifiers {
		res, err := f(tmp)
		if err != nil {
			return nil, nil, err
		}
		switch t := res.(type) {
		case RequestWrapper:
			tmp = t
		case ResponseWrapper:
			resp = new(Response)
			resp.Data = t.Data()
			resp.IsComplete = t.IsComplete()
			resp.Io = t.Io()
			resp.Metadata = Metadata{}
			resp.Metadata.Headers = t.Headers()
			resp.Metadata.StatusCode = t.StatusCode()
			break
		default:
			continue
		}
	}

	req.Method = tmp.Method()
	req.URL = tmp.URL()
	req.Query = tmp.Query()
	req.Path = tmp.Path()
	req.Body = tmp.Body()
	req.Params = tmp.Params()
	req.Headers = tmp.Headers()

	return req, resp, nil
}

func executeResponseModifiers(ctx context.Context, respModifiers []func(interface{}) (interface{}, error), r *Response, req RequestWrapper) (*Response, error) {
	var tmp ResponseWrapper
	tmp = responseWrapper{
		ctx:        ctx,
		request:    req,
		data:       r.Data,
		isComplete: r.IsComplete,
		metadata: metadataWrapper{
			headers:    r.Metadata.Headers,
			statusCode: r.Metadata.StatusCode,
		},
		io: r.Io,
	}

	for _, f := range respModifiers {
		res, err := f(tmp)
		if err != nil {
			return nil, err
		}
		t, ok := res.(ResponseWrapper)
		if !ok {
			continue
		}
		tmp = t
	}

	r.Data = tmp.Data()
	r.IsComplete = tmp.IsComplete()
	r.Io = tmp.Io()
	r.Metadata = Metadata{}
	r.Metadata.Headers = tmp.Headers()
	r.Metadata.StatusCode = tmp.StatusCode()
	return r, nil
}

// RequestWrapper is an interface for passing proxy request between the lura pipe and the loaded plugins
type RequestWrapper interface {
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
}

// ResponseWrapper is an interface for passing proxy response between the lura pipe and the loaded plugins
type ResponseWrapper interface {
	Data() map[string]interface{}
	Io() io.Reader
	IsComplete() bool
	Headers() map[string][]string
	StatusCode() int
}

func newRequestWrapper(ctx context.Context, r *Request) *requestWrapper {
	return &requestWrapper{
		ctx:     ctx,
		method:  r.Method,
		url:     r.URL,
		query:   r.Query,
		path:    r.Path,
		body:    r.Body,
		params:  r.Params,
		headers: r.Headers,
	}
}

type requestWrapper struct {
	ctx     context.Context
	method  string
	url     *url.URL
	query   url.Values
	path    string
	body    io.ReadCloser
	params  map[string]string
	headers map[string][]string
}

func (r *requestWrapper) Context() context.Context     { return r.ctx }
func (r *requestWrapper) Method() string               { return r.method }
func (r *requestWrapper) URL() *url.URL                { return r.url }
func (r *requestWrapper) Query() url.Values            { return r.query }
func (r *requestWrapper) Path() string                 { return r.path }
func (r *requestWrapper) Body() io.ReadCloser          { return r.body }
func (r *requestWrapper) Params() map[string]string    { return r.params }
func (r *requestWrapper) Headers() map[string][]string { return r.headers }

type metadataWrapper struct {
	headers    map[string][]string
	statusCode int
}

func (m metadataWrapper) Headers() map[string][]string { return m.headers }
func (m metadataWrapper) StatusCode() int              { return m.statusCode }

type responseWrapper struct {
	ctx        context.Context
	request    interface{}
	data       map[string]interface{}
	isComplete bool
	metadata   metadataWrapper
	io         io.Reader
}

func (r responseWrapper) Context() context.Context     { return r.ctx }
func (r responseWrapper) Request() interface{}         { return r.request }
func (r responseWrapper) Data() map[string]interface{} { return r.data }
func (r responseWrapper) IsComplete() bool             { return r.isComplete }
func (r responseWrapper) Io() io.Reader                { return r.io }
func (r responseWrapper) Headers() map[string][]string { return r.metadata.headers }
func (r responseWrapper) StatusCode() int              { return r.metadata.statusCode }
