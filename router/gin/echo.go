// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type echoResponse struct {
	Uri         string              `json:"req_uri"`
	UriDetails  map[string]string   `json:"req_uri_details"`
	Method      string              `json:"req_method"`
	Querystring map[string][]string `json:"req_querystring"`
	Body        string              `json:"req_body"`
	Headers     map[string][]string `json:"req_headers"`
}

// EchoHandler creates a dummy handler function, useful for quick integration tests
func EchoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body string
		if c.Request.Body != nil {
			tmp, _ := io.ReadAll(c.Request.Body)
			c.Request.Body.Close()
			body = string(tmp)
		}
		resp := echoResponse{
			Uri: c.Request.RequestURI,
			UriDetails: map[string]string{
				"user":     c.Request.URL.User.String(),
				"host":     c.Request.Host,
				"path":     c.Request.URL.Path,
				"query":    c.Request.URL.Query().Encode(),
				"fragment": c.Request.URL.Fragment,
			},
			Method:      c.Request.Method,
			Querystring: c.Request.URL.Query(),
			Body:        body,
			Headers:     c.Request.Header,
		}

		c.JSON(http.StatusOK, resp)
	}
}
