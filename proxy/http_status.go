package proxy

import (
	"context"
	"io/ioutil"
	"net/http"
)

// HTTPStatusHandler defines how we tread the http response code
type HTTPStatusHandler func(context.Context, *http.Response) (*http.Response, error)

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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, HTTPResponseError{
			Code: resp.StatusCode,
			Msg:  string(body),
		}
	}

	return resp, nil
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
