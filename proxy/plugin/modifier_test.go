//go:build integration || !race
// +build integration !race

// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/luraproject/lura/v2/logging"
)

func ExampleLoadWithLoggerAndContext() {
	var data []byte

	buf := bytes.NewBuffer(data)
	logger, err := logging.NewLogger("DEBUG", buf, "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	total, err := LoadWithLoggerAndContext(context.Background(), "./tests", ".so", RegisterModifier, logger)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if total != 2 {
		fmt.Printf("unexpected number of loaded plugins!. have %d, want 2\n", total)
		return
	}

	modFactory, ok := GetRequestModifier("lura-request-modifier-example-request")
	if !ok {
		fmt.Println("modifier factory not found in the register")
		return
	}

	modifier := modFactory(map[string]interface{}{})

	input := requestWrapper{
		ctx:     context.WithValue(context.Background(), "myCtxKey", "some"),
		path:    "/bar",
		method:  "GET",
		headers: map[string][]string{"X-Foo": {"bar"}},
	}

	tmp, err := modifier(input)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	output, ok := tmp.(RequestWrapper)
	if !ok {
		fmt.Println("unexpected result type")
		return
	}

	if res := output.Path(); res != "/bar/fooo" {
		fmt.Printf("unexpected result path. have %s, want /bar/fooo\n", res)
		return
	}

	modFactory, ok = GetResponseModifier("lura-request-modifier-example-response")
	if !ok {
		fmt.Println("modifier factory not found in the register")
		return
	}

	modifier = modFactory(map[string]interface{}{})

	response := responseWrapper{
		ctx:     context.WithValue(context.Background(), "myCtxKey", "other"),
		request: input,
		data:    map[string]interface{}{"foo": "bar"},
	}

	if _, err = modifier(response); err != nil {
		fmt.Println(err.Error())
		return
	}

	lines := strings.Split(buf.String(), "\n")
	for i := range lines[:len(lines)-1] {
		fmt.Println(lines[i][21:])
	}

	// output:
	// DEBUG: [PLUGIN: lura-error-example] Logger loaded
	// DEBUG: [PLUGIN: lura-request-modifier-example] Logger loaded
	// DEBUG: [PLUGIN: lura-request-modifier-example] Context loaded
	// DEBUG: [PLUGIN: lura-request-modifier-example] Request modifier injected
	// DEBUG: context key: some
	// DEBUG: params: map[]
	// DEBUG: headers: map[X-Foo:[bar]]
	// DEBUG: method: GET
	// DEBUG: url: <nil>
	// DEBUG: query: map[]
	// DEBUG: path: /bar/fooo
	// DEBUG: [PLUGIN: lura-request-modifier-example] Response modifier injected
	// DEBUG: Header X-Foo value: bar
	// DEBUG: context key: other
	// DEBUG: data: map[foo:bar]
	// DEBUG: is complete: false
	// DEBUG: headers: map[]
	// DEBUG: status code: 0
	// DEBUG: original headers: map[X-Foo:[bar]]

}

func TestLoad(t *testing.T) {
	total, err := LoadWithLogger("./tests", ".so", RegisterModifier, logging.NoOp)
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

	input := requestWrapper{ctx: context.WithValue(context.Background(), "myCtxKey", "some"), path: "/bar"}

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
