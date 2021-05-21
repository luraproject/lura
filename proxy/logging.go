// SPDX-License-Identifier: Apache-2.0
package proxy

import (
	"context"
	"time"

	"github.com/luraproject/lura/logging"
)

// NewLoggingMiddleware creates proxy middleware for logging requests and responses
func NewLoggingMiddleware(logger logging.Logger, name string) Middleware {
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			begin := time.Now()
			logger.Info(name, "Calling backend")
			logger.Debug("Request", request)

			result, err := next[0](ctx, request)

			logger.Info(name, "Call to backend took", time.Since(begin).String())
			if err != nil {
				logger.Warning(name, "Call to backend failed:", err.Error())
				return result, err
			}
			if result == nil {
				logger.Warning(name, "Call to backend returned a null response")
			}

			return result, err
		}
	}
}
