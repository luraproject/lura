//go:build integration || !race
// +build integration !race

// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"io"
	"net/url"
	"testing"
)

func TestLoad(t *testing.T) {
	total, err := Load("./tests", ".so", RegisterModifier)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if total != 2 {
		t.Errorf("unexpected number of loaded plugins!. have %d, want 2", total)
	}

	modFactory, ok := GetRequestModifier("lura-request-modifier-example-request")
	if !ok {
		t.Error("modifier factory not found in the register")
		return
	}

	modifier := modFactory(map[string]interface{}{})

	input := requestWrapper{path: "/bar"}

	tmp, err := modifier(input)
	if err != nil {
		t.Error(err.Error())
		return
	}

	output, ok := tmp.(RequestWrapper)
	if !ok {
		t.Error("unexpected result type")
		return
	}

	if res := output.Path(); res != "/bar/fooo" {
		t.Errorf("unexpected result path. have %s, want /bar/fooo", res)
	}
}

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
