package proxy

import (
	"context"
	"strings"
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

			parts := make(chan *Response, totalBackends)
			failed := make(chan error, totalBackends)

			for _, n := range next {
				go requestPart(localCtx, n, request, parts, failed)
			}

			acc := newIncrementalMergeAccumulator(totalBackends, combiner)
			for i := 0; i < totalBackends; i++ {
				select {
				case err := <-failed:
					acc.Merge(nil, err)
				case response := <-parts:
					acc.Merge(response, nil)
				}
			}

			result, err := acc.Result()
			cancel()
			return result, err
		}
	}
}

type incrementalMergeAccumulator struct {
	pending  int
	data     *Response
	combiner ResponseCombiner
	errs     []error
}

func newIncrementalMergeAccumulator(total int, combiner ResponseCombiner) *incrementalMergeAccumulator {
	return &incrementalMergeAccumulator{
		pending:  total,
		combiner: combiner,
		errs:     []error{},
	}
}

func (i *incrementalMergeAccumulator) Merge(res *Response, err error) {
	i.pending--
	if err != nil {
		i.errs = append(i.errs, err)
		if i.data != nil {
			i.data.IsComplete = false
		}
		return
	}
	if res == nil {
		i.errs = append(i.errs, errNullResult)
		return
	}
	if i.data == nil {
		i.data = res
		return
	}
	i.data = i.combiner(2, []*Response{i.data, res})
}

func (i *incrementalMergeAccumulator) Result() (*Response, error) {
	if i.data == nil {
		return &Response{Data: make(map[string]interface{}, 0), IsComplete: false}, newMergeError(i.errs)
	}

	if i.pending != 0 || len(i.errs) != 0 {
		i.data.IsComplete = false
	}
	return i.data, newMergeError(i.errs)
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

func newMergeError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return mergeError{errs}
}

type mergeError struct {
	errs []error
}

func (m mergeError) Error() string {
	msg := make([]string, len(m.errs))
	for i, err := range m.errs {
		msg[i] = err.Error()
	}
	return strings.Join(msg, "\n")
}

// ResponseCombiner func to merge the collected responses into a single one
type ResponseCombiner func(int, []*Response) *Response

// RegisterResponseCombiner adds a new response combiner into the internal register
func RegisterResponseCombiner(name string, f ResponseCombiner) {
	responseCombiners.SetResponseCombiner(name, f)
}

const (
	mergeKey            = "combiner"
	defaultCombinerName = "default"
)

var responseCombiners = initResponseCombiners()

func initResponseCombiners() *combinerRegister {
	return newCombinerRegister(map[string]ResponseCombiner{defaultCombinerName: combineData}, combineData)
}

func getResponseCombiner(extra config.ExtraConfig) ResponseCombiner {
	combiner, _ := responseCombiners.GetResponseCombiner(defaultCombinerName)
	if v, ok := extra[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if v, ok := e[mergeKey]; ok {
				if c, ok := responseCombiners.GetResponseCombiner(v.(string)); ok {
					combiner = c
				}
			}
		}
	}
	return combiner
}

func combineData(total int, parts []*Response) *Response {
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
