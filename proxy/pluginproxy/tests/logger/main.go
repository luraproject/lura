// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
)

func main() {}

var ProxyRegisterer = registerer("logger")

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

var (
	logger Logger          = nil
	ctx    context.Context = context.Background()
)

type registerer string

func (r registerer) RegisterProxies(f func(name string, handler Handler)) {
	f(string(r), r.registerProxies)
}

func (registerer) RegisterLogger(in interface{}) {
	l, ok := in.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ProxyRegisterer))
}

func (registerer) RegisterContext(c context.Context) {
	ctx = c
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Context loaded", ProxyRegisterer))
}

func (registerer) registerProxies(globalCtx context.Context, cfg map[string]interface{}, next ProxyWrapper) ProxyWrapper {
	// check the config
	// return the proxies

	// Graceful shutdown of any service or connection managed by the plugin
	go func() {
		<-ctx.Done()
		logger.Debug("Shuting down the service")
	}()

	if logger == nil {
		fmt.Println("request modifier loaded without logger")
		return next
	}

	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Request modifier injected", ProxyRegisterer))
	return func(ctx context.Context, rw RequestWrapper) (ResponseWrapper, error) {
		logger.Debug("params:", rw.Params())
		logger.Debug("headers:", rw.Headers())
		logger.Debug("method:", rw.Method())
		logger.Debug("url:", rw.URL())
		logger.Debug("query:", rw.Query())
		logger.Debug("path:", rw.Path())

		return next(ctx, rw)
	}
}

var unkownTypeErr = errors.New("unknown request type")

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}
