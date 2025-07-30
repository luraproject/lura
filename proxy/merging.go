// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

// NewMergeDataMiddleware creates proxy middleware for merging responses from several backends
func NewMergeDataMiddleware(logger logging.Logger, endpointConfig *config.EndpointConfig) Middleware { // skipcq: GO-R1005
	totalBackends := len(endpointConfig.Backend)
	if totalBackends == 0 {
		logger.Fatal("all endpoints must have at least one backend: NewMergeDataMiddleware")
		return nil
	}
	if totalBackends == 1 {
		return emptyMiddlewareFallback(logger)
	}
	serviceTimeout := time.Duration(85*endpointConfig.Timeout.Nanoseconds()/100) * time.Nanosecond
	combiner := getResponseCombiner(endpointConfig.ExtraConfig)
	isSequential, sequentialReplacements := sequentialMergerConfig(endpointConfig)

	logger.Debug(
		fmt.Sprintf(
			"[ENDPOINT: %s][Merge] Backends: %d, sequential: %t, combiner: %s",
			endpointConfig.Endpoint,
			totalBackends,
			isSequential,
			getResponseCombinerName(endpointConfig.ExtraConfig),
		),
	)

	bfFactory := backendFiltererFactory.filtererFactory

	return func(next ...Proxy) Proxy {
		if len(next) != totalBackends {
			// we leave the panic here, because we do not want to continue
			// if this configuration is wrong, as it would lead to unexpected
			// behaviour.
			logger.Fatal("not enough proxies for this endpoint: NewMergeDataMiddleware")
			return nil
		}
		reqClone := func(r *Request) *Request { res := r.Clone(); return &res }

		filters, err := bfFactory(endpointConfig)
		if err != nil {
			logger.Error(fmt.Sprintf("[ENDPOINT: %s][Merge] Error creating backend filterers: %s", endpointConfig.Endpoint, err))
		}

		if hasUnsafeBackends(endpointConfig) {
			reqClone = CloneRequest
		}

		if !isSequential {
			return parallelMerge(reqClone, serviceTimeout, combiner, filters, next...)
		}

		return sequentialMerge(reqClone, serviceTimeout, combiner, sequentialReplacements, filters, next...)
	}
}

type BackendFiltererFactory func(*config.EndpointConfig) ([]BackendFilterer, error)

type BackendFilterer func(*Request) bool

func defaultBackendFiltererFactory(_ *config.EndpointConfig) ([]BackendFilterer, error) {
	return []BackendFilterer{}, nil
}

type backendFiltererRegistry struct {
	filtererFactory BackendFiltererFactory
	once            *sync.Once
}

var backendFiltererFactory = backendFiltererRegistry{
	filtererFactory: defaultBackendFiltererFactory,
	once:            new(sync.Once),
}

func RegisterBackendFiltererFactory(f BackendFiltererFactory) {
	backendFiltererFactory.once.Do(func() {
		backendFiltererFactory.filtererFactory = f
	})
}

type sequentialBackendReplacement struct {
	backendIndex int
	destination  string
	source       []string
	fullResponse bool
}

func sequentialMergerConfig(cfg *config.EndpointConfig) (bool, [][]sequentialBackendReplacement) {
	enabled := false
	totalBackends := len(cfg.Backend)
	sequentialReplacements := make([][]sequentialBackendReplacement, totalBackends)
	var propagatedParams []string

	if v, ok := cfg.ExtraConfig[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if v, ok := e[isSequentialKey]; ok {
				c, ok := v.(bool)
				enabled = ok && c
			}
			if v, ok := e[sequentialPropagateKey]; ok {
				if a, ok := v.([]interface{}); ok {
					for _, p := range a {
						propagatedParams = append(propagatedParams, p.(string))
					}
				}
			}
		}
	}
	var rePropagatedParams = regexp.MustCompile(`[Rr]esp(\d+)_?([\w-.]+)?`)
	var reUrlPatterns = regexp.MustCompile(`\{\{\.Resp(\d+)_([\w-.]+)\}\}`)
	destKeyGenerator := func(i string, t string) string {
		key := "Resp" + i
		if t != "" {
			key += "_" + t
		}
		return key
	}

	for i, b := range cfg.Backend {
		for _, match := range reUrlPatterns.FindAllStringSubmatch(b.URLPattern, -1) {
			if len(match) > 1 {
				backendIndex, err := strconv.Atoi(match[1])
				if err != nil {
					continue
				}

				sequentialReplacements[i] = append(sequentialReplacements[i], sequentialBackendReplacement{
					backendIndex: backendIndex,
					destination:  destKeyGenerator(match[1], match[2]),
					source:       strings.Split(match[2], "."),
					fullResponse: match[2] == "",
				})
			}
		}

		if i > 0 {
			for _, p := range propagatedParams {
				for _, match := range rePropagatedParams.FindAllStringSubmatch(p, -1) {
					if len(match) > 1 {
						backendIndex, err := strconv.Atoi(match[1])
						if err != nil || backendIndex >= totalBackends {
							continue
						}

						sequentialReplacements[i] = append(sequentialReplacements[i], sequentialBackendReplacement{
							backendIndex: backendIndex,
							destination:  destKeyGenerator(match[1], match[2]),
							source:       strings.Split(match[2], "."),
							fullResponse: match[2] == "",
						})
					}
				}
			}
		}
	}
	return enabled, sequentialReplacements
}

func hasUnsafeBackends(cfg *config.EndpointConfig) bool {
	if len(cfg.Backend) == 1 {
		return false
	}

	for _, b := range cfg.Backend {
		if m := strings.ToUpper(b.Method); m != http.MethodGet && m != http.MethodHead {
			return true
		}
	}

	return false
}

func parallelMerge(
	reqCloner func(*Request) *Request,
	timeout time.Duration,
	rc ResponseCombiner,
	filters []BackendFilterer,
	next ...Proxy,
) Proxy {
	return func(ctx context.Context, request *Request) (*Response, error) {
		localCtx, cancel := context.WithTimeout(ctx, timeout)

		proxyCount := len(next)
		filterCount := len(filters)

		parts := make(chan *Response, proxyCount)
		failed := make(chan error, proxyCount)

		for i, n := range next {
			if (i < filterCount) && (filters[i] != nil) && !filters[i](request) {
				proxyCount--
				continue
			}
			go requestPart(localCtx, n, reqCloner(request), parts, failed)
		}

		acc := newIncrementalMergeAccumulator(proxyCount, rc)
		for i := 0; i < proxyCount; i++ {
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

func sequentialMerge(
	reqCloner func(*Request) *Request,
	timeout time.Duration,
	rc ResponseCombiner,
	sequentialReplacements [][]sequentialBackendReplacement,
	filters []BackendFilterer,
	next ...Proxy,
) Proxy { // skipcq: GO-R1005
	return func(ctx context.Context, request *Request) (*Response, error) {
		localCtx, cancel := context.WithTimeout(ctx, timeout)

		filterCount := len(filters)
		parts := make([]*Response, len(next))
		out := make(chan *Response, 1)
		errCh := make(chan error, 1)
		sequentialMergeRegistry := map[string]string{}

		acc := newIncrementalMergeAccumulator(len(next), rc)
	TxLoop:
		for i, n := range next {
			if (i < filterCount) && (filters[i] != nil) && !filters[i](request) {
				parts[i] = &Response{IsComplete: true, Data: make(map[string]interface{})}
				acc.pending--
				continue
			}

			if i > 0 {
				for _, r := range sequentialReplacements[i] {
					if r.backendIndex >= i || parts[r.backendIndex] == nil {
						continue
					}

					var v interface{}
					var ok bool

					data := parts[r.backendIndex].Data
					if len(r.source) > 1 {
						for _, k := range r.source[:len(r.source)-1] {
							v, ok = data[k]
							if !ok {
								break
							}
							clean, ok := v.(map[string]interface{})
							if !ok {
								break
							}
							data = clean
						}
					}

					if found := sequentialMergeRegistry[r.destination]; found != "" {
						request.Params[r.destination] = found
						continue
					}

					if r.fullResponse {
						if parts[r.backendIndex].Io == nil {
							continue
						}
						buf, err := io.ReadAll(parts[r.backendIndex].Io)

						if err == nil {
							request.Params[r.destination] = string(buf)
							sequentialMergeRegistry[r.destination] = string(buf)
						}
						continue
					}

					v, ok = data[r.source[len(r.source)-1]]
					if !ok {
						continue
					}

					var param string

					switch clean := v.(type) {
					case []interface{}:
						if len(clean) == 0 {
							request.Params[r.destination] = ""
							break
						}
						var b strings.Builder
						for i := 0; i < len(clean)-1; i++ {
							fmt.Fprintf(&b, "%v,", clean[i])
						}
						fmt.Fprintf(&b, "%v", clean[len(clean)-1])
						param = b.String()
					case string:
						param = clean
					case int:
						param = strconv.Itoa(clean)
					case float64:
						param = strconv.FormatFloat(clean, 'E', -1, 32)
					case bool:
						param = strconv.FormatBool(clean)
					default:
						param = fmt.Sprintf("%v", v)
					}
					request.Params[r.destination] = param
					sequentialMergeRegistry[r.destination] = param
				}
			}

			sequentialRequestPart(localCtx, n, reqCloner(request), out, errCh)

			select {
			case err := <-errCh:
				if i == 0 {
					cancel()
					return nil, err
				}
				acc.Merge(nil, err)
				break TxLoop
			case response := <-out:
				acc.Merge(response, nil)
				if !response.IsComplete {
					break TxLoop
				}
				parts[i] = response
			}
		}

		result, err := acc.Result()
		cancel()
		return result, err
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
		return nil, newMergeError(i.errs)
	}

	if i.pending > 0 || len(i.errs) > 0 {
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

func sequentialRequestPart(ctx context.Context, next Proxy, request *Request, out chan<- *Response, failed chan<- error) {
	copyRequest := CloneRequest(request)

	in, err := next(ctx, request)

	*request = *copyRequest

	if err != nil {
		failed <- err
		return
	}
	if in == nil {
		failed <- errNullResult
		return
	}
	select {
	case out <- in:
	case <-ctx.Done():
		failed <- ctx.Err()
	}
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

func (m mergeError) Errors() []error {
	return m.errs
}

// ResponseCombiner func to merge the collected responses into a single one
type ResponseCombiner func(int, []*Response) *Response

// RegisterResponseCombiner adds a new response combiner into the internal register
func RegisterResponseCombiner(name string, f ResponseCombiner) {
	responseCombiners.SetResponseCombiner(name, f)
}

const (
	mergeKey               = "combiner"
	isSequentialKey        = "sequential"
	sequentialPropagateKey = "sequential_propagated_params"
	defaultCombinerName    = "default"
)

var responseCombiners = initResponseCombiners()

func initResponseCombiners() *combinerRegister {
	return newCombinerRegister(map[string]ResponseCombiner{defaultCombinerName: combineData}, combineData)
}

func getResponseCombinerName(extra config.ExtraConfig) string {
	if v, ok := extra[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if v, ok := e[mergeKey]; ok {
				if _, ok := responseCombiners.GetResponseCombiner(v.(string)); ok {
					return v.(string)
				}
			}
		}
	}
	return defaultCombinerName
}

func getResponseCombiner(extra config.ExtraConfig) ResponseCombiner {
	combiner := getResponseCombinerName(extra)
	c, _ := responseCombiners.GetResponseCombiner(combiner)
	return c
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
			retResponse = &Response{Data: part.Data, IsComplete: isComplete}
			continue
		}
		for k, v := range part.Data {
			retResponse.Data[k] = v
		}
	}

	if nil == retResponse {
		// do not allow nil data in the response:
		return &Response{Data: make(map[string]interface{}), IsComplete: isComplete}
	}
	retResponse.IsComplete = isComplete
	return retResponse
}
