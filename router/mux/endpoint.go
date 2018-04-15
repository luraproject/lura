package mux

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
)

// HandlerFactory creates a handler function that adapts the mux router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

// EndpointHandler is a HandlerFactory that adapts the mux router with the injected proxy
// and the default RequestBuilder
var EndpointHandler = CustomEndpointHandler(NewRequest)

// CustomEndpointHandler returns a HandlerFactory with the received RequestBuilder using the default ToHTTPError function
func CustomEndpointHandler(rb RequestBuilder) HandlerFactory {
	return CustomEndpointHandlerWithHTTPError(rb, router.DefaultToHTTPError)
}

// CustomEndpointHandlerWithHTTPError returns a HandlerFactory with the received RequestBuilder
func CustomEndpointHandlerWithHTTPError(rb RequestBuilder, errF router.ToHTTPError) HandlerFactory {
	return func(configuration *config.EndpointConfig, prxy proxy.Proxy) http.HandlerFunc {
		endpointTimeout := time.Duration(configuration.Timeout) * time.Millisecond
		cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds()))
		isCacheEnabled := configuration.CacheTTL.Seconds() != 0
		render := getRender(configuration)

		headersToSend := configuration.HeadersToPass
		if len(headersToSend) == 0 {
			headersToSend = router.HeadersToSend
		}

		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(core.KrakendHeaderName, core.KrakendHeaderValue)
			if r.Method != configuration.Method {
				w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
				http.Error(w, "", http.StatusMethodNotAllowed)
				return
			}

			requestCtx, cancel := context.WithTimeout(context.Background(), endpointTimeout)

			response, err := prxy(requestCtx, rb(r, configuration.QueryString, headersToSend))
			if err != nil {
				w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
				http.Error(w, err.Error(), errF(err))
				cancel()
				return
			}

			select {
			case <-requestCtx.Done():
				w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
				http.Error(w, router.ErrInternalError.Error(), http.StatusInternalServerError)
				cancel()
				return
			default:
			}

			if response != nil {
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
				w.Header().Set(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
			}

			render(w, response)
			cancel()
		}
	}
}

// RequestBuilder is a function that creates a proxy.Request from the received http request
type RequestBuilder func(r *http.Request, queryString, headersToSend []string) *proxy.Request

// ParamExtractor is a function that extracts thes query params from the requested uri
type ParamExtractor func(r *http.Request) map[string]string

// NewRequest is a RequestBuilder that creates a proxy request from the received http request without
// processing the uri params
var NewRequest = NewRequestBuilder(func(_ *http.Request) map[string]string {
	return map[string]string{}
})

// NewRequestBuilder gets a RequestBuilder with the received ParamExtractor as a query param
// extraction mecanism
func NewRequestBuilder(paramExtractor ParamExtractor) RequestBuilder {
	return func(r *http.Request, queryString, headersToSend []string) *proxy.Request {
		params := paramExtractor(r)
		headers := make(map[string][]string, 2+len(headersToSend))
		headers["X-Forwarded-For"] = []string{r.RemoteAddr}
		headers["User-Agent"] = router.UserAgentHeaderValue

		for _, k := range headersToSend {
			if h, ok := r.Header[k]; ok {
				headers[k] = h
			}
		}

		query := make(map[string][]string, len(queryString))
		for i := range queryString {
			if v := r.URL.Query().Get(queryString[i]); v != "" {
				query[queryString[i]] = []string{v}
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
