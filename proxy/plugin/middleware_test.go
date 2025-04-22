//go:build integration || !race
// +build integration !race

// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/luraproject/lura/v2/logging"
)

func ExampleLoadMiddlewares() {
	var data []byte

	buf := bytes.NewBuffer(data)
	logger, err := logging.NewLogger("DEBUG", buf, "")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	total, err := LoadMiddlewares(context.Background(), "./tests", ".so", RegisterMiddleware, logger)
	if err != nil {
		for _, errLine := range strings.Split(err.Error(), "\n") {
			fmt.Println("'" + errLine + "'")
		}
	}
	if total != 1 {
		fmt.Printf("unexpected number of loaded plugins!. have %d, want 1\n", total)
		return
	}

	mwFactory, ok := GetMiddleware("middleware-plugin-demo")
	if !ok {
		fmt.Println("modifier factory not found in the register")
		return
	}

	input := requestWrapper{
		ctx:    context.WithValue(context.Background(), "myCtxKey", "some"),
		path:   "/bar",
		method: "GET",
	}

	modifier := mwFactory(map[string]interface{}{}, func(ctx context.Context, r interface{}) (interface{}, error) {
		rw, ok := r.(RequestWrapper)
		if !ok {
			fmt.Println("unexpected request type")
			return nil, nil
		}
		if path := rw.Path(); path != "/bar/fooo" {
			fmt.Printf("unexpected path. have %s, want /bar/fooo\n", path)
		}
		return responseWrapper{isComplete: true, data: map[string]interface{}{"foo": "bar"}}, nil
	})

	tmp, err := modifier(context.Background(), input)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	output, ok := tmp.(ResponseWrapper)
	if !ok {
		fmt.Println("unexpected result type")
		return
	}

	fmt.Printf("%+v\n", output.Data())

	lines := strings.Split(buf.String(), "\n")
	for i := range lines[:len(lines)-1] {
		fmt.Println(lines[i][21:])
	}

	// output:
	// 'plugin loader found 2 error(s): '
	// 'plugin #0 (tests/lura-error-example.so): plugin: symbol MiddlewareRegisterer not found in plugin github.com/luraproject/lura/v2/proxy/plugin/tests/error'
	// 'plugin #2 (tests/lura-request-modifier-example.so): plugin: symbol MiddlewareRegisterer not found in plugin github.com/luraproject/lura/v2/proxy/plugin/tests/logger'
	// map[extra:true foo:bar]
	// DEBUG: [PLUGIN: middleware-plugin-demo] Logger loaded
	// DEBUG: [PLUGIN: middleware-plugin-demo] Context loaded
	// DEBUG: [PLUGIN: middleware-plugin-demo] Middleware injected

}

func TestLoadMiddlewares(t *testing.T) {
	total, err := LoadMiddlewares(context.Background(), "./tests", ".so", RegisterMiddleware, logging.NoOp)
	if err == nil {
		t.Error("an error was expected")
		return
	}

	expectedErrorMsg := `plugin loader found 2 error(s): 
plugin #0 (tests/lura-error-example.so): plugin: symbol MiddlewareRegisterer not found in plugin github.com/luraproject/lura/v2/proxy/plugin/tests/error
plugin #2 (tests/lura-request-modifier-example.so): plugin: symbol MiddlewareRegisterer not found in plugin github.com/luraproject/lura/v2/proxy/plugin/tests/logger`

	if errMsg := err.Error(); errMsg != expectedErrorMsg {
		t.Errorf("unexpected error: %s", errMsg)
	}

	if total != 1 {
		t.Errorf("unexpected number of loaded plugins!. have %d, want 1", total)
	}

	mwFactory, ok := GetMiddleware("middleware-plugin-demo")
	if !ok {
		t.Error("middleware factory not found in the register")
		return
	}

	var wasNextExecuted bool

	modifier := mwFactory(map[string]interface{}{}, func(ctx context.Context, r interface{}) (interface{}, error) {
		wasNextExecuted = true
		rw, ok := r.(RequestWrapper)
		if !ok {
			t.Error("unexpected result type")
			return nil, nil
		}
		if path := rw.Path(); path != "/bar/fooo" {
			t.Errorf("unexpected result path. have %s, want /bar/fooo", path)
		}
		return responseWrapper{isComplete: true, data: map[string]interface{}{"foo": "bar"}}, nil
	})

	req := requestWrapper{ctx: context.WithValue(context.Background(), "myCtxKey", "some"), path: "/bar"}

	tmp, err := modifier(context.Background(), req)
	if err != nil {
		t.Error(err.Error())
		return
	}

	if !wasNextExecuted {
		t.Error("the next middleware wasn't executed")
	}

	output, ok := tmp.(ResponseWrapper)
	if !ok {
		t.Error("unexpected result type")
		return
	}

	data := output.Data()
	if v, ok := data["extra"].(bool); !ok || !v {
		t.Error("wrong extra field in the response data")
	}
	if v, ok := data["foo"].(string); !ok || v != "bar" {
		t.Error("wrong foo field in the response data")
	}
}

type ResponseWrapper interface {
	Data() map[string]interface{}
	Io() io.Reader
	IsComplete() bool
	Headers() map[string][]string
	StatusCode() int
}

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
