package chi

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
	"github.com/go-chi/chi"
)

const requestParamsAsterisk string = "*"

// HandlerFactory creates a handler function that adapts the chi router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

// EndpointHandler implements the HandleFactory interface using the default ToHTTPError function
func EndpointHandler(configuration *config.EndpointConfig, proxy proxy.Proxy) http.HandlerFunc {
	return CustomEndpointHandlerWithHTTPError(configuration, proxy, router.DefaultToHTTPError)
}

// CustomEndpointHandlerWithHTTPError returns a HandlerFactory with the received RequestBuilder
func CustomEndpointHandlerWithHTTPError(configuration *config.EndpointConfig, proxy proxy.Proxy, errF router.ToHTTPError) http.HandlerFunc {
	render := getRender(configuration)
	requestBuilder := NewRequestBuilder(configuration)
	cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds()))
	isCacheEnabled := configuration.CacheTTL.Seconds() != 0

	method := strings.ToTitle(configuration.Method)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(core.KrakendHeaderName, core.KrakendHeaderValue)
		if r.Method != method {
			w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}
		requestCtx, cancel := context.WithTimeout(context.Background(), configuration.Timeout)

		response, err := proxy(requestCtx, requestBuilder(r))

		select {
		case <-requestCtx.Done():
			if err == nil {
				err = router.ErrInternalError
			}
		default:
		}

		if response != nil && len(response.Data) > 0 {
			if response.IsComplete {
				w.Header().Set(router.CompleteResponseHeaderName, router.HeaderCompleteResponseValue)
				if isCacheEnabled {
					w.Header().Set("Cache-Control", cacheControlHeaderValue)
				}
			} else {
				w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
			}

			for k, v := range response.Metadata.Headers {
				w.Header().Set(k, v[0])
			}
		} else {
			if err != nil {
				w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
				http.Error(w, err.Error(), errF(err))
				cancel()
				return
			}
			w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
		}

		render(w, response)
		cancel()
	}
}

// NewRequestBuilder is a function that create a proxy.Request from the configuration
func NewRequestBuilder(config *config.EndpointConfig) func(r *http.Request) *proxy.Request {
	var re = regexp.MustCompile(`^\[?([\d.:]+)\]?(:[\d]*)$`)
	queryString := config.QueryString
	query := make(map[string][]string, len(queryString))

	return func(r *http.Request) *proxy.Request {
		params := extractParamsFromEndpoint(r, config.Endpoint)

		headers := make(map[string][]string, 2+len(config.HeadersToPass))
		for _, k := range config.HeadersToPass {
			if k == requestParamsAsterisk {
				headers = r.Header

				break
			}

			if h, ok := r.Header[k]; ok {
				headers[k] = h
			}
		}

		matches := re.FindAllStringSubmatch(r.RemoteAddr, -1)

		if len(matches) > 0 && len(matches[0]) > 1 {
			headers["X-Forwarded-For"] = []string{matches[0][1]}
		} else {
			headers["X-Forwarded-For"] = []string{r.RemoteAddr}
		}
		headers["User-Agent"] = router.UserAgentHeaderValue

		queryValues := r.URL.Query()
		for _, q := range queryString {
			if q == requestParamsAsterisk {
				query = queryValues

				break
			}

			if v, ok := queryValues[q]; ok && len(v) > 0 {
				query[q] = v
			}
		}

		return &proxy.Request{
			Method:  r.Method,
			Query:   query,
			Body:    r.Body,
			Params:  params,
			Headers: headers,
		}
	}
}

func extractParamsFromEndpoint(r *http.Request, endpoint string) map[string]string {
	ctx := r.Context()
	rctx := chi.RouteContext(ctx)

	params := map[string]string{}
	if len(rctx.URLParams.Keys) > 0 {
		for _, param := range rctx.URLParams.Keys {
			params[strings.Title(param)] = chi.URLParam(r, param)
		}
	}
	return params
}
