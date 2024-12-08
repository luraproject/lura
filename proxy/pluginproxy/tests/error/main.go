// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
)

func main() {}

var ProxyRegisterer = registerer("error")

type (
	Handler        = func(context.Context, map[string]interface{}, ProxyWrapper) ProxyWrapper
	ProxyWrapper   = func(context.Context, RequestWrapper) (ResponseWrapper, error)
	RequestWrapper = interface {
		Params() map[string]string
		Headers() map[string][]string
		Body() io.ReadCloser
		Method() string
		URL() *url.URL
		Query() url.Values
		Path() string
	}
	ResponseWrapper = interface {
		Data() map[string]interface{}
		Io() io.Reader
		IsComplete() bool
		Headers() map[string][]string
		StatusCode() int
	}
)

type registerer string

func (r registerer) RegisterProxies(f func(name string, handler Handler)) {
	f(string(r), r.registerProxies)
}

func (registerer) registerProxies(context.Context, map[string]interface{}, ProxyWrapper) ProxyWrapper {
	return func(ctx context.Context, rw RequestWrapper) (ResponseWrapper, error) {
		return nil, requestErr
	}
}

type customError struct {
	error
	statusCode int
}

func (r customError) StatusCode() int { return r.statusCode }

var requestErr = customError{
	error:      errors.New("request rejected just because"),
	statusCode: http.StatusTeapot,
}
