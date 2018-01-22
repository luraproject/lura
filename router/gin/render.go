package gin

import (
	"io"
	"net/http"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/gin-gonic/gin"
)

type Render func(*gin.Context, *proxy.Response)

func getRender(cfg *config.EndpointConfig) Render {
	if len(cfg.Backend) == 1 && cfg.Backend[0].Encoding == encoding.NOOP {
		switch cfg.Backend[0].Encoding {
		case encoding.NOOP:
			return noopRender
		case encoding.STRING:
			return stringRender
		}
	}
	return negotiatedRender
}

func negotiatedRender(c *gin.Context, response *proxy.Response) {
	switch c.NegotiateFormat(gin.MIMEJSON, gin.MIMEPlain, gin.MIMEXML) {
	case gin.MIMEXML:
		renderResponse(c.XML, response)
	case gin.MIMEPlain:
		renderResponse(c.YAML, response)
	default:
		renderResponse(c.JSON, response)
	}
}

func renderResponse(render func(int, interface{}), response *proxy.Response) {
	if response == nil {
		render(http.StatusOK, emptyResponse)
		return
	}
	render(http.StatusOK, response.Data)
}

func yamlRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.YAML(http.StatusOK, emptyResponse)
		return
	}
	c.YAML(http.StatusOK, response.Data)
}

func stringRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.String(http.StatusOK, "")
		return
	}
	if d, ok := response.Data["content"].(string); ok {
		c.String(http.StatusOK, d)
		return
	}
	c.String(http.StatusOK, "")
}

func jsonRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.JSON(http.StatusOK, emptyResponse)
		return
	}
	c.JSON(http.StatusOK, response.Data)
}

func noopRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(response.Metadata.StatusCode)
	io.Copy(c.Writer, response.Io)
}

var emptyResponse = gin.H{}
