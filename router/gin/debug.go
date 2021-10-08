// SPDX-License-Identifier: Apache-2.0
package gin

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"

	"github.com/luraproject/lura/v2/logging"
)

const logPrefix = "[ENDPOINT: /__debug/*]"


// DebugHandler creates a dummy handler function, useful for quick integration tests
func DebugHandler(logger logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug(logPrefix, "Method:", c.Request.Method)
		logger.Debug(logPrefix, "URL:", c.Request.RequestURI)
		logger.Debug(logPrefix, "Query:", c.Request.URL.Query())
		logger.Debug(logPrefix, "Params:", c.Params)
		logger.Debug(logPrefix, "Headers:", c.Request.Header)
		body, _ := ioutil.ReadAll(c.Request.Body)
		c.Request.Body.Close()
		logger.Debug(logPrefix, "Body:", string(body))
		c.JSON(200, gin.H{
			"message": "pong",
		})
	}
}
