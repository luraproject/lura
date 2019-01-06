package proxy

import (
	"bytes"
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
)

var (
	extraCfg = config.ExtraConfig{
		"shadow": true,
	}
	badExtra = config.ExtraConfig{
		"shadow": "string",
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
	request := &Request{}
	p(context.Background(), request)
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
	request := &Request{}
	_, err = p(context.Background(), request)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	if atomic.LoadUint64(&counter) != 2 {
		t.Errorf("The shadow proxy should have been called 2 times, not %d", counter)
	}
}
