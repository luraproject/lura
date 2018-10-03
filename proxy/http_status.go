package proxy

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/devopsfaith/krakend/config"
)

// HTTPStatusHandler defines how we tread the http response code
type HTTPStatusHandler func(context.Context, *http.Response) (*http.Response, error)

func getHTTPStatusHandler(remote *config.Backend) HTTPStatusHandler {
	if e, ok := remote.ExtraConfig[Namespace]; ok {
		if m, ok := e.(map[string]interface{}); ok {
			if v, ok := m["return_error_details"]; ok {
				if b, ok := v.(bool); ok && b {
					return DetailedHTTPStatusHandler
				}
			}
		}
	}
	return DefaultHTTPStatusHandler
}

// DefaultHTTPStatusHandler is the default implementation of HTTPStatusHandler
func DefaultHTTPStatusHandler(ctx context.Context, resp *http.Response) (*http.Response, error) {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, ErrInvalidStatusCode
	}

	return resp, nil
}

// NoOpHTTPStatusHandler is a NO-OP implementation of HTTPStatusHandler
func NoOpHTTPStatusHandler(_ context.Context, resp *http.Response) (*http.Response, error) {
	return resp, nil
}

// DetailedHTTPStatusHandler is a HTTPStatusHandler implementation
func DetailedHTTPStatusHandler(ctx context.Context, resp *http.Response) (*http.Response, error) {
	if r, err := DefaultHTTPStatusHandler(ctx, resp); err == nil {
		return r, nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		body = []byte{}
	}
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return resp, HTTPResponseError{
		Code: resp.StatusCode,
		Msg:  string(body),
	}
}

type HTTPResponseError struct {
	Code int    `json:"http_status_code"`
	Msg  string `json:"http_body,omitempty"`
}

func (r HTTPResponseError) Error() string {
	return r.Msg
}

func (r HTTPResponseError) StatusCode() int {
	return r.Code
}

func (r HTTPResponseError) Response() *Response {
	return &Response{
		Data:     map[string]interface{}{"error": r},
		Metadata: Metadata{StatusCode: r.Code},
	}
}
