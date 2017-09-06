package streaming

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging/gologging"
	"github.com/devopsfaith/krakend/proxy"
	"io"
)

func TestStreamDefaultFactory_noBackends(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := gologging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}
	factory := StreamDefaultFactory(logger)

	extraConfig := make(map[string]interface{})
	extraConfig["Forward"] = true

	endpointNoBackends := config.EndpointConfig{
		Backend:     []*config.Backend{},
		ExtraConfig: extraConfig,
	}

	if _, err := factory.New(&endpointNoBackends); err != proxy.ErrNoBackends {
		t.Errorf("Expecting ErrNoBackends. Got: %v\n", err)
	}
}

func TestStreamDefaultFactory_tooManyBackends(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := gologging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}
	factory := StreamDefaultFactory(logger)

	expectedMethod := "SOME"
	expectedPath := "/foo"

	backend := config.Backend{
		URLPattern: expectedPath,
		Method:     expectedMethod,
	}

	backend2 := config.Backend{
		URLPattern: expectedPath,
		Method:     expectedMethod,
	}

	extraConfig := make(map[string]interface{})
	extraConfig["Forward"] = true

	endpointTwoBackends := config.EndpointConfig{
		Backend:     []*config.Backend{&backend, &backend2},
		ExtraConfig: extraConfig,
	}

	if _, err := factory.New(&endpointTwoBackends); err != proxy.ErrTooManyBackends {
		t.Errorf("Expecting ErrTooManyBackends. Got: %v\n", err)
	}
}

func TestNewStreamDefaultFactory_ok(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := gologging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	expectedMethod := "SOME"
	expectedHost := "http://example.com/"
	expectedPath := "/foo"
	
	factory := StreamDefaultFactory(logger)

	extraConfig := make(map[string]interface{})
	extraConfig["Forward"] = true

	backend := config.Backend{
		URLPattern: expectedPath,
		Method:     expectedMethod,
	}
	endpointSingle := config.EndpointConfig{
		Backend:     []*config.Backend{&backend},
		ExtraConfig: extraConfig,
	}
	endpointMulti := config.EndpointConfig{
		Backend:         []*config.Backend{&backend, &backend},
		ConcurrentCalls: 3,
		ExtraConfig:     extraConfig,
	}
	serviceConfig := config.ServiceConfig{
		Version:   1,
		Endpoints: []*config.EndpointConfig{&endpointSingle, &endpointMulti},
		Timeout:   100 * time.Millisecond,
		Host:      []string{expectedHost},
	}
	if err := serviceConfig.Init(); err != nil {
		t.Errorf("Error during the config init: %s\n", err.Error())
	}

	proxyMulti, err := factory.New(&endpointMulti)
	if proxyMulti != nil && err != nil {
		t.Errorf("The factory returned an unexpected error: %s\n", err.Error())
	}

	proxySingle, err := factory.New(&endpointSingle)
	if proxySingle != nil && err != nil {
		t.Errorf("The factory returned an unexpected error: %s\n", err.Error())
	}

}

func newDummyReadCloser(content string) io.ReadCloser {
	return dummyReadCloser{strings.NewReader(content)}
}

type dummyReadCloser struct {
	reader io.Reader
}

func (d dummyReadCloser) Read(p []byte) (int, error) {
	return d.reader.Read(p)
}

func (d dummyReadCloser) Close() error {
	return nil
}
