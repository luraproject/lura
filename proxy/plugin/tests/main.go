// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
)

func main() {}

var ModifierRegisterer = registerer("lura-request-modifier-example")

var logger Logger = nil

type registerer string

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r), r.modifierFactory, true, false)
}

func (registerer) RegisterLogger(in interface{}) {
	l, ok := in.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ModifierRegisterer))

}

func (registerer) modifierFactory(
	map[string]interface{},
) func(interface{}) (interface{}, error) {
	// check the config
	// return the modifier

	if logger == nil {
		return func(input interface{}) (interface{}, error) {
			req, ok := input.(RequestWrapper)
			if !ok {
				return nil, unkownTypeErr
			}

			return modifier(req), nil
		}
	}

	return func(input interface{}) (interface{}, error) {
		req, ok := input.(RequestWrapper)
		if !ok {
			return nil, unkownTypeErr
		}

		r := modifier(req)

		logger.Debug("params:", r.params)
		logger.Debug("headers:", r.headers)
		logger.Debug("method:", r.method)
		logger.Debug("url:", r.url)
		logger.Debug("query:", r.query)
		logger.Debug("path:", r.path)

		return r, nil
	}
}

func modifier(req RequestWrapper) requestWrapper {
	return requestWrapper{
		params:  req.Params(),
		headers: req.Headers(),
		body:    req.Body(),
		method:  req.Method(),
		url:     req.URL(),
		query:   req.Query(),
		path:    path.Join(req.Path(), "/fooo"),
	}
}

var unkownTypeErr = errors.New("unknow request type")

type RequestWrapper interface {
	Params() map[string]string
	Headers() map[string][]string
	Body() io.ReadCloser
	Method() string
	URL() *url.URL
	Query() url.Values
	Path() string
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

func (r requestWrapper) Method() string               { return r.method }
func (r requestWrapper) URL() *url.URL                { return r.url }
func (r requestWrapper) Query() url.Values            { return r.query }
func (r requestWrapper) Path() string                 { return r.path }
func (r requestWrapper) Body() io.ReadCloser          { return r.body }
func (r requestWrapper) Params() map[string]string    { return r.params }
func (r requestWrapper) Headers() map[string][]string { return r.headers }

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}
