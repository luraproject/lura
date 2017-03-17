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

var ErrInternalError = errors.New("internal server error")

// HandlerFactory creates a handler function that adapts the mux router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

// EndpointHandler creates a handler function that adapts the net/http mux router with the injected proxy
func EndpointHandler(configuration *config.EndpointConfig, proxy proxy.Proxy) http.HandlerFunc {
	endpointTimeout := time.Duration(configuration.Timeout) * time.Millisecond

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != configuration.Method {
			http.Error(w, "", http.StatusMethodNotAllowed)
			return
		}

		requestCtx, cancel := context.WithTimeout(context.Background(), endpointTimeout)

		w.Header().Set(core.KrakendHeaderName, core.KrakendHeaderValue)

		response, err := proxy(requestCtx, NewRequest(r, configuration.QueryString))
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

var (
	headersToSend        = []string{"Content-Type"}
	userAgentHeaderValue = []string{core.KrakendUserAgent}
)

// NewRequest gets a proxy request from the received http request
func NewRequest(r *http.Request, queryString []string) *proxy.Request {
	// params := make(map[string]string, len(c.Params))
	// for _, param := range c.Params {
	// 	params[strings.Title(param.Key)] = param.Value
	// }

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
		Params:  map[string]string{},
		Headers: headers,
	}
}
