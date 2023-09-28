// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

// NewConcurrentMiddleware creates a proxy middleware that enables sending several requests concurrently
func NewConcurrentMiddleware(remote *config.Backend) Middleware {
	l, _ := logging.NewLogger("DEBUG", os.Stdout, "[LURA]")
	if remote.ConcurrentCalls == 1 {
		l.Error("TooFewConcurrentCalls: NewConcurrentMiddleware expects more than 1 concurrent call")
	}
	serviceTimeout := time.Duration(75*remote.Timeout.Nanoseconds()/100) * time.Nanosecond

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			l.Error("ErrTooManyProxies: NewConcurrentMiddleware only accepts 1 proxy, got %s (extra proxies will be ignored)", len(next))
		}

		return func(ctx context.Context, request *Request) (*Response, error) {
			localCtx, cancel := context.WithTimeout(ctx, serviceTimeout)

			results := make(chan *Response, remote.ConcurrentCalls)
			failed := make(chan error, remote.ConcurrentCalls)

			for i := 0; i < remote.ConcurrentCalls; i++ {
				go processConcurrentCall(localCtx, next[0], request, results, failed)
			}

			var response *Response
			var err error

			for i := 0; i < remote.ConcurrentCalls; i++ {
				select {
				case response = <-results:
					if response != nil && response.IsComplete {
						cancel()
						return response, nil
					}
				case err = <-failed:
				case <-ctx.Done():
				}
			}
			cancel()
			return response, err
		}
	}
}

var errNullResult = errors.New("invalid response")

func processConcurrentCall(ctx context.Context, next Proxy, request *Request, out chan<- *Response, failed chan<- error) {
	localCtx, cancel := context.WithCancel(ctx)

	result, err := next(localCtx, request)
	if err != nil {
		failed <- err
		cancel()
		return
	}
	if result == nil {
		failed <- errNullResult
		cancel()
		return
	}
	select {
	case out <- result:
	case <-ctx.Done():
		failed <- ctx.Err()
	}
	cancel()
}
