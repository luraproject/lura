// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

// NewFilterHeadersMiddleware returns a middleware with or without a header filtering
// proxy wrapping the next element (depending on the configuration).
func NewFilterHeadersMiddleware(_ logging.Logger, remote *config.Backend) Middleware {
	if len(remote.HeadersToPass) == 0 {
		return EmptyMiddleware
	}

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		nextProxy := next[0]
		return func(ctx context.Context, request *Request) (*Response, error) {
			if len(request.Headers) == 0 {
				return nextProxy(ctx, request)
			}
			numHeadersToPass := 0
			for _, v := range remote.HeadersToPass {
				if _, ok := request.Headers[v]; ok {
					numHeadersToPass++
				}
			}
			if numHeadersToPass == len(request.Headers) {
				// all the headers should pass, no need to clone the headers
				return nextProxy(ctx, request)
			}
			// ATTENTION: this is not a clone of headers!
			// this just filters the headers we do not want to send:
			// issues and race conditions could happen the same way as when we
			// do not filter the headers. This is a design decission, and if we
			// want to clone the header values (because of write modifications),
			// that should be done at an upper level (so the approach is the same
			// for non filtered parallel requests).
			newHeaders := make(map[string][]string, numHeadersToPass)
			for _, v := range remote.HeadersToPass {
				if values, ok := request.Headers[v]; ok {
					newHeaders[v] = values
				}
			}
			return nextProxy(ctx, &Request{
				Method:  request.Method,
				URL:     request.URL,
				Query:   request.Query,
				Path:    request.Path,
				Body:    request.Body,
				Params:  request.Params,
				Headers: newHeaders,
			})
		}
	}
}
