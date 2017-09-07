package config

import (
	"strings"
	"testing"
	"time"
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
		subject := ServiceConfig{Version: 1, Endpoints: []*EndpointConfig{{Endpoint: e}}}
		err := subject.Init()
		if err == nil || strings.Index(err.Error(), "ERROR: the endpoint url path [") != 0 {
			t.Error("Error expected processing", e)
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
	endpoint := EndpointConfig{Backend: []*Backend{&backend}}
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

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
	subject := ServiceConfig{Endpoints: []*EndpointConfig{&endpoint}, uriParser: NewURIParser()}

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

func TestConfig_init(t *testing.T) {
	supuBackend := Backend{
		URLPattern: "/__debug/supu",
	}
	supuEndpoint := EndpointConfig{
		Endpoint: "/supu",
		Method:   "post",
		Timeout:  1500 * time.Millisecond,
		CacheTTL: 6 * time.Hour,
		Backend:  []*Backend{&supuBackend},
	}

	githubBackend := Backend{
		URLPattern: "/",
		Host:       []string{"https://api.github.com"},
		Whitelist:  []string{"authorizations_url", "code_search_url"},
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
		Version:   1,
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

	if supuBackend.Method != "POST" {
		t.Error("supuBackend method not sanitized")
	}

	if userBackend.Timeout != subject.Timeout {
		t.Error("default timeout not applied to the userBackend")
	}

	if userEndpoint.CacheTTL != subject.CacheTTL {
		t.Error("default CacheTTL not applied to the userEndpoint")
	}
}

func TestConfig_initKONoBackends(t *testing.T) {
	subject := ServiceConfig{
		Version: 1,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/supu",
				Method:   "post",
				Backend:  []*Backend{},
			},
		},
	}

	if err := subject.Init(); err == nil ||
		!strings.HasPrefix(err.Error(), "WARNING: the [/supu] endpoint has 0 backends defined! Ignoring") {
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
		Version: 1,
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
		Version: 1,
		Host:    []string{"http://127.0.0.1:8080"},
		Endpoints: []*EndpointConfig{
			{
				Endpoint: "/__debug/supu",
				Method:   "get",
				Backend:  []*Backend{},
			},
		},
	}

	if err := subject.Init(); err == nil ||
		!strings.HasPrefix(err.Error(), "error parsing regexp: missing closing ): `a(b`") {
		t.Error("Expecting an error at the configuration init!", err)
	}

	debugPattern = dp
}

func TestDefaultConfigGetter(t *testing.T) {
	getter, ok := ConfigGetters[defaultNamespace]
	if !ok {
		t.Error("Nothing stored at the default namespace")
		return
	}
	extra := ExtraConfig{
		"a": 1,
		"b": true,
		"c": []int{1, 1, 2, 3, 5, 8},
	}
	result := getter(extra)
	res, ok := result.(ExtraConfig)
	if !ok {
		t.Error("error casting the returned value")
		return
	}
	if v, ok := res["a"]; !ok || 1 != v.(int) {
		t.Errorf("unexpected value for key `a`: %v", v)
		return
	}
	if v, ok := res["b"]; !ok || !v.(bool) {
		t.Errorf("unexpected value for key `b`: %v", v)
		return
	}
	if v, ok := res["c"]; !ok || 6 != len(v.([]int)) {
		t.Errorf("unexpected value for key `c`: %v", v)
		return
	}
}

func TestConfigGetter(t *testing.T) {
	ConfigGetters = map[string]ConfigGetter{
		"ns1": func(e ExtraConfig) interface{} {
			return len(e)
		},
		"ns2": func(e ExtraConfig) interface{} {
			v, ok := e["publishedAt"]
			if !ok {
				return e
			}
			start, ok := v.(time.Time)
			if !ok {
				return e
			}
			if start.After(time.Now()) {
				return nil
			}
			return e
		},
	}
	extra := ExtraConfig{
		"a": 1,
		"b": true,
		"c": []int{1, 1, 2, 3, 5, 8},
	}

	tmp1 := ConfigGetters["ns1"](extra)
	v, ok := tmp1.(int)
	if !ok {
		t.Error("error casting the returned value")
		return
	}
	if 3 != v {
		t.Errorf("unexpected value for getter `ns1`: %v", v)
		return
	}

	tmp2 := ConfigGetters["ns2"](extra)
	res, ok := tmp2.(ExtraConfig)
	if !ok {
		t.Error("error casting the returned value")
		return
	}
	if v, ok := res["a"]; !ok || 1 != v.(int) {
		t.Errorf("unexpected value for key `a`: %v", v)
		return
	}
	if v, ok := res["b"]; !ok || !v.(bool) {
		t.Errorf("unexpected value for key `b`: %v", v)
		return
	}
	if v, ok := res["c"]; !ok || 6 != len(v.([]int)) {
		t.Errorf("unexpected value for key `c`: %v", v)
		return
	}
}
