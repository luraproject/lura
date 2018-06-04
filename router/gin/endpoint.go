package gin

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router"
)

// HandlerFactory creates a handler function that adapts the gin router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) gin.HandlerFunc

// EndpointHandler implements the HandleFactory interface using the default ToHTTPError function
func EndpointHandler(configuration *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
	return CustomErrorEndpointHandler(configuration, proxy, router.DefaultToHTTPError)
}

// CustomErrorEndpointHandler implements the HandleFactory interface
func CustomErrorEndpointHandler(configuration *config.EndpointConfig, prxy proxy.Proxy, errF router.ToHTTPError) gin.HandlerFunc {
	endpointTimeout := time.Duration(configuration.Timeout) * time.Millisecond
	cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds()))
	isCacheEnabled := configuration.CacheTTL.Seconds() != 0
	requestGenerator := NewRequest(configuration.HeadersToPass)
	render := getRender(configuration)

	return func(c *gin.Context) {
		requestCtx, cancel := context.WithTimeout(c, endpointTimeout)

		c.Header(core.KrakendHeaderName, core.KrakendHeaderValue)

		response, err := prxy(requestCtx, requestGenerator(c, configuration.QueryString))
		if err != nil {
			abort(c, errF(err))
			cancel()
			return
		}

		select {
		case <-requestCtx.Done():
			abort(c, http.StatusInternalServerError)
			cancel()
			return
		default:
		}

		if response != nil {
			if response.IsComplete {
				c.Header(router.CompleteResponseHeaderName, router.HeaderCompleteResponseValue)
				if isCacheEnabled {
					c.Header("Cache-Control", cacheControlHeaderValue)
				}
			} else {
				c.Header(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
			}

			for k, v := range response.Metadata.Headers {
				c.Header(k, v[0])
			}
		} else {
			c.Header(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
		}

		render(c, response)
		cancel()
	}
}

func abort(c *gin.Context, code int) {
	c.Status(code)
	c.Header(router.CompleteResponseHeaderName, router.HeaderIncompleteResponseValue)
	c.Abort()
}

// NewRequest gets a request from the current gin context and the received query string
func NewRequest(headersToSend []string) func(*gin.Context, []string) *proxy.Request {
	if len(headersToSend) == 0 {
		headersToSend = router.HeadersToSend
	}

	return func(c *gin.Context, queryString []string) *proxy.Request {
		params := make(map[string]string, len(c.Params))
		for _, param := range c.Params {
			params[strings.Title(param.Key)] = param.Value
		}

		headers := make(map[string][]string, 2+len(headersToSend))
		headers["X-Forwarded-For"] = []string{c.ClientIP()}
		headers["User-Agent"] = router.UserAgentHeaderValue

		for _, k := range headersToSend {
			if h, ok := c.Request.Header[k]; ok {
				headers[k] = h
			}
		}

		query := make(map[string][]string, len(queryString))
		for i := range queryString {
			if v := c.Request.URL.Query().Get(queryString[i]); v != "" {
				query[queryString[i]] = []string{v}
			}
		}

		return &proxy.Request{
			Method:  c.Request.Method,
			Query:   query,
			Body:    c.Request.Body,
			Params:  params,
			Headers: headers,
		}
	}
}
