package proxy

import (
	"context"
	"sync"
	"time"

	"github.com/devopsfaith/krakend/config"
)

// NewMergeDataMiddleware creates proxy middleware for merging responses from several backends
func NewMergeDataMiddleware(endpointConfig *config.EndpointConfig) Middleware {
	totalBackends := len(endpointConfig.Backend)
	if totalBackends == 0 {
		panic(ErrNoBackends)
	}
	if totalBackends == 1 {
		return EmptyMiddleware
	}
	serviceTimeout := time.Duration(85*endpointConfig.Timeout.Nanoseconds()/100) * time.Nanosecond
	combiner := getResponseCombiner(endpointConfig.ExtraConfig)

	return func(next ...Proxy) Proxy {
		if len(next) != totalBackends {
			panic(ErrNotEnoughProxies)
		}

		return func(ctx context.Context, request *Request) (*Response, error) {
			localCtx, cancel := context.WithTimeout(ctx, serviceTimeout)

			parts := make(chan *Response, len(next))
			failed := make(chan error, len(next))

			for _, n := range next {
				go requestPart(localCtx, n, request, parts, failed)
			}

			var err error
			responses := make([]*Response, len(next))
			isEmpty := true
			for i := 0; i < len(next); i++ {
				select {
				case err = <-failed:
				case responses[i] = <-parts:
					isEmpty = false
				}
			}
			if isEmpty {
				cancel()
				return &Response{Data: make(map[string]interface{}), IsComplete: false}, err
			}

			result := combiner(localCtx, totalBackends, responses)
			cancel()
			return result, err
		}
	}
}

func requestPart(ctx context.Context, next Proxy, request *Request, out chan<- *Response, failed chan<- error) {
	localCtx, cancel := context.WithCancel(ctx)

	in, err := next(localCtx, request)
	if err != nil {
		failed <- err
		cancel()
		return
	}
	if in == nil {
		failed <- errNullResult
		cancel()
		return
	}
	select {
	case out <- in:
	case <-ctx.Done():
		failed <- ctx.Err()
	}
	cancel()
}

// ResponseCombiner func to merge the collected responses into a single one
type ResponseCombiner func(context.Context, int, []*Response) *Response

// RegisterResponseCombiner adds a new response combiner into the internal register
func RegisterResponseCombiner(name string, f ResponseCombiner) {
	responseCombinersMutex.Lock()
	responseCombiners[name] = f
	responseCombinersMutex.Unlock()
}

const (
	mergeKey            = "combiner"
	defaultCombinerName = "default"
)

var (
	responseCombinersMutex = &sync.RWMutex{}
	responseCombiners      = map[string]ResponseCombiner{
		defaultCombinerName: combineData,
	}
)

func getResponseCombiner(extra config.ExtraConfig) ResponseCombiner {
	responseCombinersMutex.RLock()
	combiner := responseCombiners[defaultCombinerName]
	if v, ok := extra[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if v, ok := e[mergeKey]; ok {
				if c, ok := responseCombiners[v.(string)]; ok {
					combiner = c
				}
			}
		}
	}
	responseCombinersMutex.RUnlock()
	return combiner
}

func combineData(_ context.Context, total int, parts []*Response) *Response {
	isComplete := len(parts) == total
	var retResponse *Response
	for _, part := range parts {
		if part == nil || part.Data == nil {
			isComplete = false
			continue
		}
		isComplete = isComplete && part.IsComplete
		if retResponse == nil {
			retResponse = part
			continue
		}
		for k, v := range part.Data {
			retResponse.Data[k] = v
		}
	}

	if nil == retResponse {
		// do not allow nil data in the response:
		return &Response{Data: make(map[string]interface{}, 0), IsComplete: isComplete}
	}
	retResponse.IsComplete = isComplete
	return retResponse
}
