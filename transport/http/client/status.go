// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/luraproject/lura/v2/config"
)

// Namespace to be used in extra config
const Namespace = "github.com/devopsfaith/krakend/http"

// ErrInvalidStatusCode is the error returned by the http proxy when the received status code
// is not a 200 nor a 201
var ErrInvalidStatusCode = errors.New("invalid status code")

// HTTPStatusHandler defines how we tread the http response code
type HTTPStatusHandler func(context.Context, *http.Response) (*http.Response, error)

// GetHTTPStatusHandler returns a status handler. If the 'return_error_details' key is defined
// at the extra config, it returns a DetailedHTTPStatusHandler. Otherwise, it returns a
// DefaultHTTPStatusHandler
func GetHTTPStatusHandler(remote *config.Backend) HTTPStatusHandler {
	if e, ok := remote.ExtraConfig[Namespace]; ok {
		if m, ok := e.(map[string]interface{}); ok {
			if v, ok := m["return_error_details"]; ok {
				if b, ok := v.(string); ok && b != "" {
					return DetailedHTTPStatusHandler(b)
				}
			} else if v, ok := m["return_error_code"].(bool); ok && v {
				return ErrorHTTPStatusHandler
			}
		}
	}
	return DefaultHTTPStatusHandler
}

// DefaultHTTPStatusHandler is the default implementation of HTTPStatusHandler
func DefaultHTTPStatusHandler(_ context.Context, resp *http.Response) (*http.Response, error) {
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, ErrInvalidStatusCode
	}

	return resp, nil
}

// ErrorHTTPStatusHandler is a HTTPStatusHandler that returns the status code as part of the error details
func ErrorHTTPStatusHandler(ctx context.Context, resp *http.Response) (*http.Response, error) {
	if _, err := DefaultHTTPStatusHandler(ctx, resp); err == nil {
		return resp, nil
	}
	return resp, newHTTPResponseError(resp)
}

// NoOpHTTPStatusHandler is a NO-OP implementation of HTTPStatusHandler
func NoOpHTTPStatusHandler(_ context.Context, resp *http.Response) (*http.Response, error) {
	return resp, nil
}

// DetailedHTTPStatusHandler is a HTTPStatusHandler implementation
func DetailedHTTPStatusHandler(name string) HTTPStatusHandler {
	return func(ctx context.Context, resp *http.Response) (*http.Response, error) {
		if _, err := DefaultHTTPStatusHandler(ctx, resp); err == nil {
			return resp, nil
		}

		return resp, NamedHTTPResponseError{
			HTTPResponseError: newHTTPResponseError(resp),
			name:              name,
		}
	}
}

func newHTTPResponseError(resp *http.Response) HTTPResponseError {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		body = []byte{}
	}
	resp.Body.Close()
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	return HTTPResponseError{
		Code: resp.StatusCode,
		Msg:  string(body),
	}
}

// HTTPResponseError is the error to be returned by the ErrorHTTPStatusHandler
type HTTPResponseError struct {
	Code int    `json:"http_status_code"`
	Msg  string `json:"http_body,omitempty"`
}

// Error returns the error message
func (r HTTPResponseError) Error() string {
	return r.Msg
}

// StatusCode returns the status code returned by the backend
func (r HTTPResponseError) StatusCode() int {
	return r.Code
}

// NamedHTTPResponseError is the error to be returned by the DetailedHTTPStatusHandler
type NamedHTTPResponseError struct {
	HTTPResponseError
	name string
}

// Name returns the name of the backend where the error happened
func (r NamedHTTPResponseError) Name() string {
	return r.name
}
