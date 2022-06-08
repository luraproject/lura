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
		return EmptyMiddleware
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
		return EmptyMiddleware
	}

	return newPluginMiddleware(logger, "BACKEND", remote.URLPattern, cfg)
}

func newPluginMiddleware(logger logging.Logger, tag, pattern string, cfg map[string]interface{}) Middleware {
	plugins, ok := cfg["name"].([]interface{})
	if !ok {
		return EmptyMiddleware
	}

	reqModifiers := []func(interface{}) (interface{}, error){}
	respModifiers := []func(interface{}) (interface{}, error){}

	for _, p := range plugins {
		name, ok := p.(string)
		if !ok {
			continue
		}

		if mf, ok := plugin.GetRequestModifier(name); ok {
			reqModifiers = append(reqModifiers, mf(cfg))
			continue
		}

		if mf, ok := plugin.GetResponseModifier(name); ok {
			respModifiers = append(respModifiers, mf(cfg))
		}
	}

	totReqModifiers, totRespModifiers := len(reqModifiers), len(respModifiers)
	if totReqModifiers == totRespModifiers && totRespModifiers == 0 {
		return EmptyMiddleware
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
			panic(ErrTooManyProxies)
		}

		if totReqModifiers == 0 {
			return func(ctx context.Context, r *Request) (*Response, error) {
				resp, err := next[0](ctx, r)
				if err != nil {
					return resp, err
				}

				return executeResponseModifiers(respModifiers, resp)
			}
		}

		if totRespModifiers == 0 {
			return func(ctx context.Context, r *Request) (*Response, error) {
				var err error
				r, err = executeRequestModifiers(reqModifiers, r)
				if err != nil {
					return nil, err
				}

				return next[0](ctx, r)
			}
		}

		return func(ctx context.Context, r *Request) (*Response, error) {
			var err error
			r, err = executeRequestModifiers(reqModifiers, r)
			if err != nil {
				return nil, err
			}

			resp, err := next[0](ctx, r)
			if err != nil {
				return resp, err
			}

			return executeResponseModifiers(respModifiers, resp)
		}
	}
}

func executeRequestModifiers(reqModifiers []func(interface{}) (interface{}, error), r *Request) (*Request, error) {
	var tmp RequestWrapper
	tmp = &requestWrapper{
		method:  r.Method,
		url:     r.URL,
		query:   r.Query,
		path:    r.Path,
		body:    r.Body,
		params:  r.Params,
		headers: r.Headers,
	}

	for _, f := range reqModifiers {
		res, err := f(tmp)
		if err != nil {
			return nil, err
		}
		t, ok := res.(RequestWrapper)
		if !ok {
			continue
		}
		tmp = t
	}

	r.Method = tmp.Method()
	r.URL = tmp.URL()
	r.Query = tmp.Query()
	r.Path = tmp.Path()
	r.Body = tmp.Body()
	r.Params = tmp.Params()
	r.Headers = tmp.Headers()

	return r, nil
}

func executeResponseModifiers(respModifiers []func(interface{}) (interface{}, error), r *Response) (*Response, error) {
	var tmp ResponseWrapper
	tmp = responseWrapper{
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

type requestWrapper struct {
	method  string
	url     *url.URL
	query   url.Values
	path    string
	body    io.ReadCloser
	params  map[string]string
	headers map[string][]string
}

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
	data       map[string]interface{}
	isComplete bool
	metadata   metadataWrapper
	io         io.Reader
}

func (r responseWrapper) Data() map[string]interface{} { return r.data }
func (r responseWrapper) IsComplete() bool             { return r.isComplete }
func (r responseWrapper) Io() io.Reader                { return r.io }
func (r responseWrapper) Headers() map[string][]string { return r.metadata.headers }
func (r responseWrapper) StatusCode() int              { return r.metadata.statusCode }
