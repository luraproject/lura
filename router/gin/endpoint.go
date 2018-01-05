package gin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
	"github.com/devopsfaith/krakend/encoding"
	"github.com/devopsfaith/krakend/proxy"
)

// ErrInternalError is the error returned by the router when something went wrong
var ErrInternalError = errors.New("internal server error")

// HandlerFactory creates a handler function that adapts the gin router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) gin.HandlerFunc

// EndpointHandler implements the HandleFactory interface
func EndpointHandler(configuration *config.EndpointConfig, prxy proxy.Proxy) gin.HandlerFunc {
	endpointTimeout := time.Duration(configuration.Timeout) * time.Millisecond

	var dump func(*gin.Context, *proxy.Response)
	if len(configuration.Backend) == 1 && configuration.Backend[0].Encoding == encoding.NOOP {
		dump = noopResponse
	} else {
		dump = jsonResponse
	}

	return func(c *gin.Context) {
		requestCtx, cancel := context.WithTimeout(c, endpointTimeout)

		c.Header(core.KrakendHeaderName, core.KrakendHeaderValue)

		response, err := prxy(requestCtx, NewRequest(c, configuration.QueryString))
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			cancel()
			return
		}

		select {
		case <-requestCtx.Done():
			c.AbortWithError(http.StatusInternalServerError, ErrInternalError)
			cancel()
		default:
		}

		if configuration.CacheTTL.Seconds() != 0 && response != nil && response.IsComplete {
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds())))
		}

		if response != nil {
			for k, v := range response.Metadata.Headers {
				c.Header(k, v[0])
			}
		}

		dump(c, response)
		cancel()
	}
}

func jsonResponse(c *gin.Context, response *proxy.Response) {
	if response != nil {
		c.JSON(http.StatusOK, response.Data)
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func noopResponse(c *gin.Context, response *proxy.Response) {
	if response != nil {
		c.Status(response.Metadata.StatusCode)
		io.Copy(c.Writer, response.Io)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

var (
	headersToSend        = []string{"Content-Type"}
	userAgentHeaderValue = []string{core.KrakendUserAgent}
)

// NewRequest gets a request from the current gin context and the received query string
func NewRequest(c *gin.Context, queryString []string) *proxy.Request {
	params := make(map[string]string, len(c.Params))
	for _, param := range c.Params {
		params[strings.Title(param.Key)] = param.Value
	}

	headers := make(map[string][]string, 2+len(headersToSend))
	headers["X-Forwarded-For"] = []string{c.ClientIP()}
	headers["User-Agent"] = userAgentHeaderValue

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
