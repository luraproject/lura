package proxy

import (
	"bytes"
	"io"
	"net/url"
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
		key := []byte{}
		key = append(key, "{{."...)
		key = append(key, k...)
		key = append(key, "}}"...)
		buff = bytes.Replace(buff, key, []byte(v), -1)
	}
	r.Path = string(buff)
}

// Clone clones itself into a new request
func (r *Request) Clone() Request {
	return Request{
		Method:  r.Method,
		URL:     r.URL,
		Query:   r.Query,
		Path:    r.Path,
		Body:    r.Body,
		Params:  r.Params,
		Headers: r.Headers,
	}
}
