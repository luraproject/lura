package mux

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
	"github.com/devopsfaith/krakend/proxy"
)

// ErrInternalError is the error returned by the router when something went wrong
var ErrInternalError = errors.New("internal server error")

// HandlerFactory creates a handler function that adapts the mux router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

// EndpointHandler is a HandlerFactory that adapts the mux router with the injected proxy
// and the default RequestBuilder
var EndpointHandler = CustomEndpointHandler(NewRequest)

// CustomEndpointHandler returns a HandlerFactory with the received RequestBuilder
func CustomEndpointHandler(rb RequestBuilder) HandlerFactory {
	return func(configuration *config.EndpointConfig, proxy proxy.Proxy) http.HandlerFunc {
		endpointTimeout := time.Duration(configuration.Timeout) * time.Millisecond

		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(core.KrakendHeaderName, core.KrakendHeaderValue)
			if r.Method != configuration.Method {
				http.Error(w, "", http.StatusMethodNotAllowed)
				return
			}

			requestCtx, cancel := context.WithTimeout(context.Background(), endpointTimeout)

			response, err := proxy(requestCtx, rb(r, configuration.QueryString))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				cancel()
				return
			}

			select {
			case <-requestCtx.Done():
				http.Error(w, ErrInternalError.Error(), http.StatusInternalServerError)
				cancel()
				return
			default:
			}

			var js []byte

			if response != nil {
				js, err = json.Marshal(response.Data)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					cancel()
					return
				}
				if configuration.CacheTTL.Seconds() != 0 && response.IsComplete {
					w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds())))
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
			cancel()
		}
	}
}

// RequestBuilder is a function that creates a proxy.Request from the received http request
type RequestBuilder func(r *http.Request, queryString []string) *proxy.Request

// ParamExtractor is a function that extracts thes query params from the requested uri
type ParamExtractor func(r *http.Request) map[string]string

// NewRequest is a RequestBuilder that creates a proxy request from the received http request without
// processing the uri params
var NewRequest = NewRequestBuilder(func(r *http.Request) map[string]string {
	return map[string]string{}
})

var (
	headersToSend        = []string{"Content-Type"}
	userAgentHeaderValue = []string{core.KrakendUserAgent}
)

// NewRequestBuilder gets a RequestBuilder with the received ParamExtractor as a query param
// extraction mecanism
func NewRequestBuilder(paramExtractor ParamExtractor) RequestBuilder {
	return func(r *http.Request, queryString []string) *proxy.Request {
		params := paramExtractor(r)
		headers := make(map[string][]string, 2+len(headersToSend))
		headers["X-Forwarded-For"] = []string{r.RemoteAddr}
		headers["User-Agent"] = userAgentHeaderValue

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
