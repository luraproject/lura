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

var (
	renderRegister = map[string]Render{
		"negotiate":     negotiatedRender,
		encoding.NOOP:   noopRender,
		encoding.STRING: stringRender,
		encoding.JSON:   jsonRender,
	}
)

func RegisterRender(name string, r Render) {
	renderRegister[name] = r
}

func getRender(cfg *config.EndpointConfig) Render {
	fallback := jsonRender
	if len(cfg.Backend) > 0 {
		fallback = getRenderFromBackend(cfg.Backend[0])
	}

	if cfg.OutputEncoding == "" {
		return fallback
	}

	r, ok := renderRegister[cfg.OutputEncoding]
	if !ok {
		return fallback
	}
	return r
}

func getRenderFromBackend(cfg *config.Backend) Render {
	r, ok := renderRegister[cfg.Encoding]
	if !ok {
		return jsonRender
	}
	return r
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

func renderResponse(render func(int, interface{}), response *proxy.Response) {
	if response == nil {
		render(http.StatusOK, emptyResponse)
		return
	}
	render(http.StatusOK, response.Data)
}

var emptyResponse = gin.H{}
