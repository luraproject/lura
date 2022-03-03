// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestConfig_rejectInvalidVersion(t *testing.T) {
	subject := ServiceConfig{}
	err := subject.Init()
	if err == nil || strings.Index(err.Error(), "unsupported version: 0 (want: 3)") != 0 {
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
		subject := ServiceConfig{Version: ConfigVersion, Endpoints: []*EndpointConfig{{Endpoint: e, Method: "GET"}}}
		err := subject.Init()
		if err == nil || err.Error() != fmt.Sprintf("ignoring the 'GET %s' endpoint, since it is invalid!!!", e) {
			t.Errorf("Unexpected error processing '%s': %v", e, err)
		}
	}
}

func TestConfig_initBackendURLMappings_ok(t *testing.T) {
	samples := []string{
		"supu/{tupu}",
		"/supu/{tupu1}",
		"/supu.local/",
		"supu/{tupu_56}/{supu-5t6}?a={foo}&b={foo}",
		"supu/{tupu_56}{supu-5t6}?a={foo}&b={foo}",
		"supu/tupu{supu-5t6}?a={foo}&b={foo}",
		"{resp0_x}/{tupu1}/{tupu_56}{supu-5t6}?a={tupu}&b={foo}",
		"{resp0_x}/{tupu1}/{JWT.foo}",
		"{resp0_x}/{tupu1}/{JWT.http://example.com/foo_bar}",
	}

	expected := []string{
		"/supu/{{.Tupu}}",
		"/supu/{{.Tupu1}}",
		"/supu.local/",
		"/supu/{{.Tupu_56}}/{{.Supu-5t6}}?a={{.Foo}}&b={{.Foo}}",
		"/supu/{{.Tupu_56}}{{.Supu-5t6}}?a={{.Foo}}&b={{.Foo}}",
		"/supu/tupu{{.Supu-5t6}}?a={{.Foo}}&b={{.Foo}}",
		"/{{.Resp0_x}}/{{.Tupu1}}/{{.Tupu_56}}{{.Supu-5t6}}?a={{.Tupu}}&b={{.Foo}}",
		"/{{.Resp0_x}}/{{.Tupu1}}/{{.JWT.foo}}",
		"/{{.Resp0_x}}/{{.Tupu1}}/{{.JWT.http://example.com/foo_bar}}",
	}

	backend := Backend{}
	endpoint := EndpointConfig{Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

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
	endpoint := EndpointConfig{
		Method:   "GET",
		Endpoint: "/some/{tupu}",
		Backend:  []*Backend{&backend},
	}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

	inputSet := map[string]interface{}{
		"tupu": nil,
	}

	expectedErrMsg := "input and output params do not match. endpoint: GET /some/{tupu}, backend: 0. input: [tupu], output: [foo supu-5t6 tupu_56]"

	err := subject.initBackendURLMappings(0, 0, inputSet)
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestConfig_initBackendURLMappings_undefinedOutput(t *testing.T) {
	backend := Backend{URLPattern: "supu/{tupu_56}/{supu-5t6}?a={foo}&b={foo}"}
	endpoint := EndpointConfig{Endpoint: "/", Method: "GET", Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

	inputSet := map[string]interface{}{
		"tupu": nil,
		"supu": nil,
		"foo":  nil,
	}

	expectedErrMsg := "undefined output param 'supu-5t6'! endpoint: GET /, backend: 0. input: [foo supu tupu], output: [foo supu-5t6 tupu_56]"
	err := subject.initBackendURLMappings(0, 0, inputSet)
	if err == nil || err.Error() != expectedErrMsg {
		t.Errorf("error expected. have: %v", err)
	}
}

func TestConfig_init(t *testing.T) {
	supuBackend := Backend{
		URLPattern: "/__debug/supu",
	}
	supuEndpoint := EndpointConfig{
		Endpoint:       "/supu",
		Method:         "post",
		Timeout:        1500 * time.Millisecond,
		CacheTTL:       6 * time.Hour,
		Backend:        []*Backend{&supuBackend},
		OutputEncoding: "some_render",
	}

	githubBackend := Backend{
		URLPattern: "/",
		Host:       []string{"https://api.github.com"},
		AllowList:  []string{"authorizations_url", "code_search_url"},
	}
	githubEndpoint := EndpointConfig{
		Endpoint: "/github",
		Timeout:  1500 * time.Millisecond,
		CacheTTL: 6 * time.Hour,
		Backend:  []*Backend{&githubBackend},
	}

	userBackend := Backend{
		URLPattern: "/users/{user}",
		Host:       []string{"https://jsonplaceholder.typicode.com"},
		Mapping:    map[string]string{"email": "personal_email"},
	}
	rssBackend := Backend{
		URLPattern: "/users/{user}",
		Host:       []string{"https://jsonplaceholder.typicode.com"},
		Encoding:   "rss",
	}
	postBackend := Backend{
		URLPattern: "/posts/{user}",
		Host:       []string{"https://jsonplaceholder.typicode.com"},
		Group:      "posts",
		Encoding:   "xml",
	}
	userEndpoint := EndpointConfig{
		Endpoint: "/users/{user}",
		Backend:  []*Backend{&userBackend, &rssBackend, &postBackend},
	}

	subject := ServiceConfig{
		Version:   ConfigVersion,
		Timeout:   5 * time.Second,
		CacheTTL:  30 * time.Minute,
		Host:      []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{&supuEndpoint, &githubEndpoint, &userEndpoint},
	}

	if err := subject.Init(); err != nil {
		t.Error("Error at the configuration init:", err.Error())
	}

	if len(supuBackend.Host) != 1 || supuBackend.Host[0] != subject.Host[0] {
		t.Error("Default hosts not applied to the supu backend", supuBackend.Host)
	}

	for level, method := range map[string]string{
		"userBackend":  userBackend.Method,
		"postBackend":  postBackend.Method,
		"userEndpoint": userEndpoint.Method,
	} {
		if method != "GET" {
			t.Errorf("Default method not applied at %s. Get: %s", level, method)
		}
	}

	if supuBackend.Method != "post" {
		t.Error("unexpected supuBackend")
	}

	if userBackend.Timeout != subject.Timeout {
		t.Error("default timeout not applied to the userBackend")
	}

	if userEndpoint.CacheTTL != subject.CacheTTL {
		t.Error("default CacheTTL not applied to the userEndpoint")
	}

	hash, err := subject.Hash()
	if err != nil {
		t.Error(err.Error())
	}

	if hash != "3YYe7crlYj5Qpm/oUoBqO2mQrKcalJmAoNfkRYM7aDI=" {
		t.Errorf("unexpected hash: %s", hash)
	}
}

func TestConfig_initKONoBackends(t *testing.T) {
	subject := ServiceConfig{
		Version: ConfigVersion,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/supu",
				Method:   "POST",
				Backend:  []*Backend{},
			},
		},
	}

	if err := subject.Init(); err == nil ||
		err.Error() != "ignoring the 'POST /supu' endpoint, since it has 0 backends defined!" {
		t.Error("Unexpected error at the configuration init!", err)
	}
}

func TestConfig_initKOMultipleBackendsForNoopEncoder(t *testing.T) {
	subject := ServiceConfig{
		Version: ConfigVersion,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint:       "/supu",
				Method:         "post",
				OutputEncoding: "no-op",
				Backend: []*Backend{
					{
						Encoding: "no-op",
					},
					{
						Encoding: "no-op",
					},
				},
			},
		},
	}

	if err := subject.Init(); err != errInvalidNoOpEncoding {
		t.Error("Expecting an error at the configuration init!", err)
	}
}

func TestConfig_initKOInvalidHost(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The init process did not panic with an invalid host!")
		}
	}()
	subject := ServiceConfig{
		Version: ConfigVersion,
		Host:    []string{"http://127.0.0.1:8080http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/supu",
				Method:   "post",
				Backend:  []*Backend{},
			},
		},
	}

	subject.Init()
}

func TestConfig_initKOInvalidDebugPattern(t *testing.T) {
	dp := debugPattern

	debugPattern = "a(b"
	subject := ServiceConfig{
		Version: ConfigVersion,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/__debug/supu",
				Method:   "GET",
				Backend:  []*Backend{},
			},
		},
	}

	if err := subject.Init(); err == nil ||
		err.Error() != "ignoring the 'GET /__debug/supu' endpoint due to a parsing error: error parsing regexp: missing closing ): `a(b`" {
		t.Error("Expecting an error at the configuration init!", err)
	}

	debugPattern = dp
}
