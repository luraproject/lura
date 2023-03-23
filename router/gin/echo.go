// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DebugHandler creates a dummy handler function, useful for quick integration tests
func EchoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		c.JSON(http.StatusOK, gin.H{
			"Method":  c.Request.Method,
			"URL":     c.Request.RequestURI,
			"Query":   c.Request.URL.Query(),
			"Params":  c.Params,
			"Headers": c.Request.Header,
			"Body":    string(body),
		})
	}
}
