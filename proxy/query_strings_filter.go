// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"net/url"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

// NewFilterQueryStringsMiddleware returns a middleware with or without a header filtering
// proxy wrapping the next element (depending on the configuration).
func NewFilterQueryStringsMiddleware(logger logging.Logger, remote *config.Backend) Middleware {
	if len(remote.QueryStringsToPass) == 0 {
		return emptyMiddlewareFallback(logger)
	}

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			logger.Fatal("too many proxies for this %s -> %s proxy middleware: NewFilterQueryStringsMiddleware only accepts 1 proxy, got %d", remote.ParentEndpoint, remote.URLPattern, len(next))
			return nil
		}
		nextProxy := next[0]
		return func(ctx context.Context, request *Request) (*Response, error) {
			if len(request.Query) == 0 {
				return nextProxy(ctx, request)
			}
			numQueryStringsToPass := 0
			for _, v := range remote.QueryStringsToPass {
				if _, ok := request.Query[v]; ok {
					numQueryStringsToPass++
				}
			}
			if numQueryStringsToPass == len(request.Query) {
				// all the query strings should pass, no need to clone the headers
				return nextProxy(ctx, request)
			}
			// ATTENTION: this is not a clone of query strings!
			// this just filters the query strings we do not want to send:
			// issues and race conditions could happen the same way as when we
			// do not filter the headers. This is a design decission, and if we
			// want to clone the query string values (because of write modifications),
			// that should be done at an upper level (so the approach is the same
			// for non filtered parallel requests).
			newQueryStrings := make(url.Values, numQueryStringsToPass)
			for _, v := range remote.QueryStringsToPass {
				if values, ok := request.Query[v]; ok {
					newQueryStrings[v] = values
				}
			}
			return nextProxy(ctx, &Request{
				Method:  request.Method,
				URL:     request.URL,
				Query:   newQueryStrings,
				Path:    request.Path,
				Body:    request.Body,
				Params:  request.Params,
				Headers: request.Headers,
			})
		}
	}
}
