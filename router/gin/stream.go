package gin

import (
	"github.com/gin-gonic/gin"
	"time"
	"net/http"
	"fmt"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"context"
	"github.com/devopsfaith/krakend/core"
	"io"
)

// EndpointHandler implements the HandleFactory interface
func StreamEndpointHandler(configuration *config.EndpointConfig, proxy proxy.Proxy) gin.HandlerFunc {
	endpointTimeout := time.Duration(configuration.Timeout) * time.Millisecond

	return func(c *gin.Context) {
		requestCtx, cancel := context.WithTimeout(c, endpointTimeout)

		c.Header(core.KrakendHeaderName, core.KrakendHeaderValue)

		response, err := proxy(requestCtx, NewRequest(c, configuration.QueryString))
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
			io.Copy(c.Writer, response.Io)
			c.Status(http.StatusOK)
		}

		cancel()
	}
}
