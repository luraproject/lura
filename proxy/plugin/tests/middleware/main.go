// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
)

var MiddlewareRegisterer = registerer("middleware-plugin-demo")

var (
	logger                Logger          = nil
	ctx                   context.Context = context.Background()
	unkownRequestTypeErr                  = errors.New("unknown request type")
	unkownResponseTypeErr                 = errors.New("unknown response type")
)

type registerer string

func (r registerer) RegisterMiddlewares(f func(
	name string,
	middlewareFactory func(map[string]interface{}, func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error),
)) {
	f(string(r), r.middlewareFactory)
}

func (r registerer) middlewareFactory(cfg map[string]interface{}, next func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error) {
	// TODO: parse the config
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Middleware injected", r))
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		reqw, ok := req.(RequestWrapper)
		if !ok {
			return nil, unkownRequestTypeErr
		}

		resp, err := next(ctx, requestWrapper{
			params:  reqw.Params(),
			headers: reqw.Headers(),
			body:    reqw.Body(),
			method:  reqw.Method(),
			url:     reqw.URL(),
			query:   reqw.Query(),
			path:    reqw.Path() + "/fooo",
		})
		respw, ok := resp.(ResponseWrapper)
		if !ok {
			return nil, unkownResponseTypeErr
		}

		data := respw.Data()
		data["extra"] = true

		return responseWrapper{
			ctx:        respw.Context(),
			request:    respw.Request(),
			data:       data,
			isComplete: respw.IsComplete(),
			metadata: metadataWrapper{
				headers:    respw.Headers(),
				statusCode: respw.StatusCode(),
			},
			io: respw.Io(),
		}, err
	}
}

func (r registerer) RegisterLogger(in interface{}) {
	l, ok := in.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", r))
}

func (r registerer) RegisterContext(c context.Context) {
	ctx = c
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Context loaded", r))
}

func main() {}

type ResponseWrapper interface {
	Context() context.Context
	Request() interface{}
	Data() map[string]interface{}
	IsComplete() bool
	Headers() map[string][]string
	StatusCode() int
	Io() io.Reader
}

type RequestWrapper interface {
	Context() context.Context
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
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

func (r requestWrapper) Context() context.Context     { return r.ctx }
func (r requestWrapper) Method() string               { return r.method }
func (r requestWrapper) URL() *url.URL                { return r.url }
func (r requestWrapper) Query() url.Values            { return r.query }
func (r requestWrapper) Path() string                 { return r.path }
func (r requestWrapper) Body() io.ReadCloser          { return r.body }
func (r requestWrapper) Params() map[string]string    { return r.params }
func (r requestWrapper) Headers() map[string][]string { return r.headers }

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

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}
