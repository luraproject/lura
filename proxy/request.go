// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"io"
	"net/url"
	"sync"
)

// Request contains the data to send to the backend
type Request struct {
	Method  string
	URL     *url.URL
	Query   url.Values
	Path    string
	Body    io.ReadCloser
	Params  map[string]string
	Headers map[string][]string
}

// GeneratePath takes a pattern and updates the path of the request
func (r *Request) GeneratePath(URLPattern string) {
	if len(r.Params) == 0 {
		r.Path = URLPattern
		return
	}
	buff := []byte(URLPattern)
	for k, v := range r.Params {
		var key []byte

		key = append(key, "{{."...)
		key = append(key, k...)
		key = append(key, "}}"...)
		buff = bytes.ReplaceAll(buff, key, []byte(v))
	}
	r.Path = string(buff)
}

// Clone clones itself into a new request. The returned cloned request is not
// thread-safe, so changes on request.Params and request.Headers could generate
// race-conditions depending on the part of the pipe they are being executed.
// For thread-safe request headers and/or params manipulation, use the proxy.CloneRequest
// function.
func (r *Request) Clone() Request {
	var clonedURL *url.URL
	if r.URL != nil {
		clonedURL, _ = url.Parse(r.URL.String())
	}
	return Request{
		Method:  r.Method,
		URL:     clonedURL,
		Query:   r.Query,
		Path:    r.Path,
		Body:    r.Body,
		Params:  r.Params,
		Headers: r.Headers,
	}
}

// bufferPool is a pool of bytes.Buffer used to reduce memory allocations
// during request body cloning in the proxy pipeline.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// CloneRequest returns a deep copy of the received request, so the received and the
// returned proxy.Request do not share a pointer
func CloneRequest(r *Request) *Request {
	clone := r.Clone()
	clone.Headers = CloneRequestHeaders(r.Headers)
	clone.Params = CloneRequestParams(r.Params)
	if r.Body == nil {
		return &clone
	}
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	buf.Reset()
	buf.ReadFrom(r.Body)
	r.Body.Close()

	data := buf.Bytes()
	originalBody := make([]byte, len(data))
	copy(originalBody, data)
	clonedBody := make([]byte, len(data))
	copy(clonedBody, data)

	r.Body = io.NopCloser(bytes.NewReader(originalBody))
	clone.Body = io.NopCloser(bytes.NewReader(clonedBody))

	return &clone
}

// CloneRequestHeaders returns a copy of the received request headers
func CloneRequestHeaders(headers map[string][]string) map[string][]string {
	m := make(map[string][]string, len(headers))
	for k, vs := range headers {
		tmp := make([]string, len(vs))
		copy(tmp, vs)
		m[k] = tmp
	}
	return m
}

// CloneRequestParams returns a copy of the received request params
func CloneRequestParams(params map[string]string) map[string]string {
	m := make(map[string]string, len(params))
	for k, v := range params {
		m[k] = v
	}
	return m
}
