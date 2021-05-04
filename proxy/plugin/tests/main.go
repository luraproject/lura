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

func init() {
	fmt.Println(string(ModifierRegisterer), "loaded!!!")
}

var ModifierRegisterer = registerer("krakend-request-modifier-example")

type registerer string

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r), r.modifierFactory, true, false)
	fmt.Println(string(ModifierRegisterer), "registered!!!")
}

func (r registerer) modifierFactory(map[string]interface{}) func(interface{}) (interface{}, error) {
	// check the config
	// return the modifier
	fmt.Println(string(ModifierRegisterer), "injected!!!")
	return func(input interface{}) (interface{}, error) {
		req, ok := input.(RequestWrapper)
		if !ok {
			return nil, unkownTypeErr
		}

		return requestWrapper{
			params:  req.Params(),
			headers: req.Headers(),
			body:    req.Body(),
			method:  req.Method(),
			url:     req.URL(),
			query:   req.Query(),
			path:    path.Join(req.Path(), "/fooo"),
		}, nil
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
