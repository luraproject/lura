// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func TestRequestGeneratePath(t *testing.T) {
	r := Request{
		Method: "GET",
		Params: map[string]string{
			"Supu": "42",
			"Tupu": "false",
			"Foo":  "bar",
		},
	}

	for i, testCase := range [][]string{
		{"/a/{{.Supu}}", "/a/42"},
		{"/a?b={{.Tupu}}", "/a?b=false"},
		{"/a/{{.Supu}}/foo/{{.Foo}}", "/a/42/foo/bar"},
		{"/a", "/a"},
	} {
		r.GeneratePath(testCase[0])
		if r.Path != testCase[1] {
			t.Errorf("%d: want %s, have %s", i, testCase[1], r.Path)
		}
	}
}

func TestRequest_Clone(t *testing.T) {
	r := Request{
		Method: "GET",
		Params: map[string]string{
			"Supu": "42",
			"Tupu": "false",
			"Foo":  "bar",
		},
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
	}
	clone := r.Clone()

	if len(r.Params) != len(clone.Params) {
		t.Errorf("wrong num of params. have: %d, want: %d", len(clone.Params), len(r.Params))
		return
	}
	for k, v := range r.Params {
		if res, ok := clone.Params[k]; !ok {
			t.Errorf("param %s not cloned", k)
		} else if res != v {
			t.Errorf("unexpected param %s. have: %s, want: %s", k, res, v)
		}
	}

	if len(r.Headers) != len(clone.Headers) {
		t.Errorf("wrong num of headers. have: %d, want: %d", len(clone.Headers), len(r.Headers))
		return
	}

	for k, vs := range r.Headers {
		if res, ok := clone.Headers[k]; !ok {
			t.Errorf("header %s not cloned", k)
		} else if len(res) != len(vs) {
			t.Errorf("unexpected header %s. have: %v, want: %v", k, res, vs)
		}
	}

	r.Headers["extra"] = []string{"supu"}

	if len(r.Headers) != len(clone.Headers) {
		t.Errorf("wrong num of headers. have: %d, want: %d", len(clone.Headers), len(r.Headers))
		return
	}

	for k, vs := range r.Headers {
		if res, ok := clone.Headers[k]; !ok {
			t.Errorf("header %s not cloned", k)
		} else if len(res) != len(vs) {
			t.Errorf("unexpected header %s. have: %v, want: %v", k, res, vs)
		}
	}
}

func TestCloneRequest(t *testing.T) {
	body := `{"a":1,"b":2}`
	r := Request{
		Method: "POST",
		Params: map[string]string{
			"Supu": "42",
			"Tupu": "false",
			"Foo":  "bar",
		},
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: ioutil.NopCloser(strings.NewReader(body)),
	}
	clone := CloneRequest(&r)

	if len(r.Params) != len(clone.Params) {
		t.Errorf("wrong num of params. have: %d, want: %d", len(clone.Params), len(r.Params))
		return
	}
	for k, v := range r.Params {
		if res, ok := clone.Params[k]; !ok {
			t.Errorf("param %s not cloned", k)
		} else if res != v {
			t.Errorf("unexpected param %s. have: %s, want: %s", k, res, v)
		}
	}

	if len(r.Headers) != len(clone.Headers) {
		t.Errorf("wrong num of headers. have: %d, want: %d", len(clone.Headers), len(r.Headers))
		return
	}

	for k, vs := range r.Headers {
		if res, ok := clone.Headers[k]; !ok {
			t.Errorf("header %s not cloned", k)
		} else if len(res) != len(vs) {
			t.Errorf("unexpected header %s. have: %v, want: %v", k, res, vs)
		}
	}

	r.Headers["extra"] = []string{"supu"}

	if _, ok := clone.Headers["extra"]; ok {
		t.Error("the cloned instance shares its headers with the original one")
	}

	delete(r.Params, "Supu")

	if _, ok := clone.Params["Supu"]; !ok {
		t.Error("the cloned instance shares its params with the original one")
	}

	rb, _ := ioutil.ReadAll(r.Body)
	cb, _ := ioutil.ReadAll(clone.Body)

	if !bytes.Equal(cb, rb) || body != string(rb) {
		t.Errorf("unexpected bodies. original: %s, returned: %s", string(rb), string(cb))
	}
}
