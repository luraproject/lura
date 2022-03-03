// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

var (
	extraCfg = config.ExtraConfig{
		Namespace: map[string]interface{}{
			"shadow": true,
		},
	}
	badExtra = config.ExtraConfig{
		Namespace: map[string]interface{}{
			"shadow": "string",
		},
	}
)

func newAssertionProxy(counter *uint64) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		atomic.AddUint64(counter, 1)
		return nil, nil
	}
}

func TestIsShadowBackend(t *testing.T) {

	cfg := &config.Backend{ExtraConfig: extraCfg}
	badCfg := &config.Backend{ExtraConfig: badExtra}

	if !isShadowBackend(cfg) {
		t.Error("The shadow backend should be true")
	}

	if isShadowBackend(&config.Backend{}) {
		t.Error("The shadow backend should be false")
	}

	if isShadowBackend(badCfg) {
		t.Error("The shadow backend should be false")
	}
}

func TestShadowMiddleware(t *testing.T) {
	var counter uint64
	assertProxy := newAssertionProxy(&counter)
	p := ShadowMiddleware(assertProxy, assertProxy)
	p(context.Background(), &Request{})
	time.Sleep(100 * time.Millisecond)
	if atomic.LoadUint64(&counter) != 2 {
		t.Errorf("The shadow proxy should have been called 2 times, not %d", counter)
	}
}

func TestShadowFactory_noBackends(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := logging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}
	factory := DefaultFactory(logger)

	sFactory := NewShadowFactory(factory)

	if _, err := sFactory.New(&config.EndpointConfig{}); err != ErrNoBackends {
		t.Errorf("Expecting ErrNoBackends. Got: %v\n", err)
	}
}

func TestNewShadowFactory(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := logging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}
	var counter uint64
	assertProxy := newAssertionProxy(&counter)
	factory := NewDefaultFactory(func(_ *config.Backend) Proxy { return assertProxy }, logger)
	f := NewShadowFactory(factory)
	sBackend := &config.Backend{ExtraConfig: extraCfg}
	backend := &config.Backend{}
	endpointConfig := &config.EndpointConfig{Backend: []*config.Backend{sBackend, backend}}
	serviceConfig := config.ServiceConfig{
		Version:   config.ConfigVersion,
		Endpoints: []*config.EndpointConfig{endpointConfig},
		Timeout:   100 * time.Millisecond,
		Host:      []string{"dummy"},
	}
	if err = serviceConfig.Init(); err != nil {
		t.Errorf("Error during the config init: %s\n", err.Error())
	}

	p, err := f.New(endpointConfig)
	if err != nil {
		t.Error(err)
	}
	_, err = p(context.Background(), &Request{})
	if err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	if atomic.LoadUint64(&counter) != 2 {
		t.Errorf("The shadow proxy should have been called 2 times, not %d", counter)
	}
}

func TestShadowMiddleware_erroredBackend(t *testing.T) {
	timeout := 100 * time.Millisecond
	p := ShadowMiddleware(
		delayedProxy(t, timeout, &Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		func(_ context.Context, _ *Request) (*Response, error) {
			return nil, errors.New("ignore me")
		},
	)
	mustEnd := time.After(time.Duration(5 * timeout))
	out, err := p(context.Background(), &Request{Params: map[string]string{}})
	if err != nil {
		t.Errorf("unexpected error: %s\n", err.Error())
		return
	}
	if out == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
		if len(out.Data) != 1 {
			t.Errorf("We weren't expecting a partial response but we got %v!\n", out)
		}
		if !out.IsComplete {
			t.Errorf("We were expecting a completed response!\n")
		}
	}
}

func TestShadowMiddleware_partialTimeout(t *testing.T) {
	timeout := 200 * time.Millisecond
	p := ShadowMiddleware(
		delayedProxy(t, time.Duration(5*timeout), &Response{Data: map[string]interface{}{"supu": 42}}),
		delayedProxy(t, time.Duration(timeout/2), &Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}))
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out, err := p(ctx, &Request{})
	if err == nil || err.Error() != "context deadline exceeded" {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if out != nil {
		t.Errorf("The proxy did not return a null result: %+v\n", out)
		return
	}
}
