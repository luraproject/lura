package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func main() {}

var ModifierRegisterer = registerer("lura-shortcircuit-example")

type registerer string

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
),
) {
	f(string(r)+"-request", r.requestModifierFactory, true, false)
	f(string(r)+"-response", r.reqsponseModifierFactory, false, true)
}

func (r registerer) requestModifierFactory(_ map[string]interface{}) func(interface{}) (interface{}, error) {
	return func(input interface{}) (interface{}, error) {
		req, ok := input.(RequestWrapper)
		if !ok {
			return nil, unknownTypeErr
		}

		header := make(http.Header)
		header.Add("X-Plugin-Request", "shortcircuit")
		return responseWrapper{
			request:    req,
			io:         strings.NewReader("shortcircuit"),
			headers:    header,
			statusCode: http.StatusTeapot,
		}, nil
	}
}

func (r registerer) reqsponseModifierFactory(_ map[string]interface{}) func(interface{}) (interface{}, error) {
	return func(input interface{}) (interface{}, error) {
		resp, ok := input.(ResponseWrapper)
		if !ok {
			return nil, unknownTypeErr
		}

		header := http.Header(resp.Headers())
		header.Add("X-Plugin-Response", "shortcircuit")
		return resp, nil
	}
}

type responseWrapper struct {
	ctx        context.Context
	request    interface{}
	data       map[string]interface{}
	isComplete bool
	headers    map[string][]string
	statusCode int
	io         io.Reader
}

func (r responseWrapper) Context() context.Context     { return r.ctx }
func (r responseWrapper) Request() interface{}         { return r.request }
func (r responseWrapper) Data() map[string]interface{} { return r.data }
func (r responseWrapper) IsComplete() bool             { return r.isComplete }
func (r responseWrapper) Io() io.Reader                { return r.io }
func (r responseWrapper) Headers() map[string][]string { return r.headers }
func (r responseWrapper) StatusCode() int              { return r.statusCode }

var unknownTypeErr = errors.New("unknown request type")

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

type ResponseWrapper interface {
	Context() context.Context
	Request() interface{}
	Data() map[string]interface{}
	IsComplete() bool
	Io() io.Reader
	Headers() map[string][]string
	StatusCode() int
}
