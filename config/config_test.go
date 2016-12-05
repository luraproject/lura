package config

import (
	"strings"
	"testing"
)

func TestConfig_rejectInvalidVersion(t *testing.T) {
	subject := ServiceConfig{}
	err := subject.Init()
	if err == nil || strings.Index(err.Error(), "Unsupported version: 0") != 0 {
		t.Error("Error expected. Got", err.Error())
	}
}

func TestConfig_rejectInvalidEndpoints(t *testing.T) {
	samples := []string{
		"/__debug",
		"/__debug/",
		"/__debug/foo",
		"/__debug/foo/bar",
	}

	for _, e := range samples {
		subject := ServiceConfig{Version: 1, Endpoints: []*EndpointConfig{&EndpointConfig{Endpoint: e}}}
		err := subject.Init()
		if err == nil || strings.Index(err.Error(), "ERROR: the endpoint url path [") != 0 {
			t.Error("Error expected processing", e)
		}
	}
}

func TestConfig_cleanHosts(t *testing.T) {
	samples := []string{
		"supu",
		"127.0.0.1",
		"https://supu.local/",
		"http://127.0.0.1",
		"supu_42.local:8080/",
		"http://127.0.0.1:8080",
	}

	expected := []string{
		"http://supu",
		"http://127.0.0.1",
		"https://supu.local",
		"http://127.0.0.1",
		"http://supu_42.local:8080",
		"http://127.0.0.1:8080",
	}

	subject := ServiceConfig{}
	result := subject.cleanHosts(samples)
	for i := range result {
		if expected[i] != result[i] {
			t.Errorf("want: %s, have: %s\n", expected[i], result[i])
		}
	}
}

func TestConfig_cleanPath(t *testing.T) {
	samples := []string{
		"supu/{tupu}",
		"/supu/{tupu}",
		"/supu.local/",
		"supu_supu.txt",
		"supu_42.local?a=8080",
		"supu/supu/supu?a=1&b=2",
		"debug/supu/supu?a=1&b=2",
	}

	expected := []string{
		"/supu/{tupu}",
		"/supu/{tupu}",
		"/supu.local/",
		"/supu_supu.txt",
		"/supu_42.local?a=8080",
		"/supu/supu/supu?a=1&b=2",
		"/debug/supu/supu?a=1&b=2",
	}

	subject := ServiceConfig{}

	for i := range samples {
		if have := subject.cleanPath(samples[i]); expected[i] != have {
			t.Errorf("want: %s, have: %s\n", expected[i], have)
		}
	}
}

func TestConfig_getEndpointPath(t *testing.T) {
	samples := []string{
		"supu/{tupu}",
		"/supu/{tupu}",
		"/supu.local/",
		"supu/{tupu}/{supu}?a={s}&b=2",
	}

	expected := []string{
		"supu/:tupu",
		"/supu/:tupu",
		"/supu.local/",
		"supu/:tupu/:supu?a={s}&b=2",
	}

	subject := ServiceConfig{}

	for i := range samples {
		params := subject.extractPlaceHoldersFromURLTemplate(samples[i], endpointURLKeysPattern)
		if have := subject.getEndpointPath(samples[i], params); expected[i] != have {
			t.Errorf("want: %s, have: %s\n", expected[i], have)
		}
	}
}

func TestConfig_initBackendURLMappings_ok(t *testing.T) {
	samples := []string{
		"supu/{tupu}",
		"/supu/{tupu1}",
		"/supu.local/",
		"supu/{tupu_56}/{supu-5t6}?a={foo}&b={foo}",
	}

	expected := []string{
		"/supu/{{.Tupu}}",
		"/supu/{{.Tupu1}}",
		"/supu.local/",
		"/supu/{{.Tupu_56}}/{{.Supu-5t6}}?a={{.Foo}}&b={{.Foo}}",
	}

	backend := Backend{}
	endpoint := EndpointConfig{Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}}

	inputSet := map[string]interface{}{
		"tupu":     nil,
		"tupu1":    nil,
		"tupu_56":  nil,
		"supu-5t6": nil,
		"foo":      nil,
	}

	for i := range samples {
		backend.URLPattern = samples[i]
		if err := subject.initBackendURLMappings(0, 0, inputSet); err != nil {
			t.Error(err)
		}
		if backend.URLPattern != expected[i] {
			t.Errorf("want: %s, have: %s\n", expected[i], backend.URLPattern)
		}
	}
}

func TestConfig_initBackendURLMappings_tooManyOutput(t *testing.T) {
	backend := Backend{URLPattern: "supu/{tupu_56}/{supu-5t6}?a={foo}&b={foo}"}
	endpoint := EndpointConfig{Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}}

	inputSet := map[string]interface{}{
		"tupu": nil,
	}

	err := subject.initBackendURLMappings(0, 0, inputSet)
	if err == nil || strings.Index(err.Error(), "Too many output params!") != 0 {
		t.Error("Error expected")
	}
}

func TestConfig_initBackendURLMappings_undefinedOutput(t *testing.T) {
	backend := Backend{URLPattern: "supu/{tupu_56}/{supu-5t6}?a={foo}&b={foo}"}
	endpoint := EndpointConfig{Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}}

	inputSet := map[string]interface{}{
		"tupu": nil,
		"supu": nil,
		"foo":  nil,
	}

	err := subject.initBackendURLMappings(0, 0, inputSet)
	if err == nil || strings.Index(err.Error(), "Undefined output param [") != 0 {
		t.Error("Error expected")
	}
}
