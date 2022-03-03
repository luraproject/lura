// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"context"
	"fmt"
	"net/textproto"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/core"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/transport/http/server"
)

const requestParamsAsterisk string = "*"

// HandlerFactory creates a handler function that adapts the gin router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) gin.HandlerFunc

// EndpointHandler implements the HandleFactory interface using the default ToHTTPError function
var EndpointHandler = CustomErrorEndpointHandler(logging.NoOp, server.DefaultToHTTPError)

// CustomErrorEndpointHandler returns a HandleFactory using the injected ToHTTPError function and logger
func CustomErrorEndpointHandler(logger logging.Logger, errF server.ToHTTPError) HandlerFactory {
	return func(configuration *config.EndpointConfig, prxy proxy.Proxy) gin.HandlerFunc {
		cacheControlHeaderValue := fmt.Sprintf("public, max-age=%d", int(configuration.CacheTTL.Seconds()))
		isCacheEnabled := configuration.CacheTTL.Seconds() != 0
		requestGenerator := NewRequest(configuration.HeadersToPass)
		render := getRender(configuration)
		logPrefix := "[ENDPOINT: " + configuration.Endpoint + "]"

		return func(c *gin.Context) {
			requestCtx, cancel := context.WithTimeout(c, configuration.Timeout)

			c.Header(core.KrakendHeaderName, core.KrakendHeaderValue)

			response, err := prxy(requestCtx, requestGenerator(c, configuration.QueryString))

			select {
			case <-requestCtx.Done():
				if err == nil {
					err = server.ErrInternalError
				}
			default:
			}

			complete := server.HeaderIncompleteResponseValue

			if response != nil && len(response.Data) > 0 {
				if response.IsComplete {
					complete = server.HeaderCompleteResponseValue
					if isCacheEnabled {
						c.Header("Cache-Control", cacheControlHeaderValue)
					}
				}

				for k, vs := range response.Metadata.Headers {
					for _, v := range vs {
						c.Writer.Header().Add(k, v)
					}
				}
			}

			c.Header(server.CompleteResponseHeaderName, complete)

			if err != nil {
				if t, ok := err.(multiError); ok {
					for i, errN := range t.Errors() {
						logger.Error(fmt.Sprintf("%s Error #%d: %s", logPrefix, i, errN.Error()))
					}
				} else {
					logger.Error(logPrefix, err.Error())
				}

				if response == nil {
					if t, ok := err.(responseError); ok {
						c.Status(t.StatusCode())
					} else {
						c.Status(errF(err))
					}
					if returnErrorMsg {
						c.Writer.WriteString(err.Error())
					}
					cancel()
					return
				}
			}

			render(c, response)
			cancel()
		}
	}
}

// NewRequest gets a request from the current gin context and the received query string
func NewRequest(headersToSend []string) func(*gin.Context, []string) *proxy.Request {
	if len(headersToSend) == 0 {
		headersToSend = server.HeadersToSend
	}

	return func(c *gin.Context, queryString []string) *proxy.Request {
		params := make(map[string]string, len(c.Params))
		for _, param := range c.Params {
			params[strings.Title(param.Key[:1])+param.Key[1:]] = param.Value
		}

		headers := make(map[string][]string, 3+len(headersToSend))

		for _, k := range headersToSend {
			if k == requestParamsAsterisk {
				headers = c.Request.Header

				break
			}

			if h, ok := c.Request.Header[textproto.CanonicalMIMEHeaderKey(k)]; ok {
				headers[k] = h
			}
		}

		headers["X-Forwarded-For"] = []string{c.ClientIP()}
		headers["X-Forwarded-Host"] = []string{c.Request.Host}
		// if User-Agent is not forwarded using headersToSend, we set
		// the KrakenD router User Agent value
		if _, ok := headers["User-Agent"]; !ok {
			headers["User-Agent"] = server.UserAgentHeaderValue
		} else {
			headers["X-Forwarded-Via"] = server.UserAgentHeaderValue
		}

		query := make(map[string][]string, len(queryString))
		queryValues := c.Request.URL.Query()
		for i := range queryString {
			if queryString[i] == requestParamsAsterisk {
				query = c.Request.URL.Query()

				break
			}

			if v, ok := queryValues[queryString[i]]; ok && len(v) > 0 {
				query[queryString[i]] = v
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

type responseError interface {
	error
	StatusCode() int
}

type multiError interface {
	error
	Errors() []error
}
