package proxy

import (
	"context"
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
