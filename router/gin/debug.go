// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"

	"github.com/luraproject/lura/v2/logging"
)

// DebugHandler creates a dummy handler function, useful for quick integration tests
func DebugHandler(logger logging.Logger) gin.HandlerFunc {
	logPrefixSecondary := "[ENDPOINT: /__debug/*]"
	return func(c *gin.Context) {
		logger.Debug(logPrefixSecondary, "Method:", c.Request.Method)
		logger.Debug(logPrefixSecondary, "URL:", c.Request.RequestURI)
		logger.Debug(logPrefixSecondary, "Query:", c.Request.URL.Query())
		logger.Debug(logPrefixSecondary, "Params:", c.Params)
		logger.Debug(logPrefixSecondary, "Headers:", c.Request.Header)
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		logger.Debug(logPrefixSecondary, "Body:", string(body))
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
}
