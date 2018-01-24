package gin

import (
	"io"
	"net/http"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/gin-gonic/gin"
)

// Render defines the signature of the functions to be use for the final response
// encoding and rendering
type Render func(*gin.Context, *proxy.Response)

// NEGOTIATE defines the value of the OutputEncoding for the negotiated render
const NEGOTIATE = "negotiate"

var (
	renderRegister = map[string]Render{
		NEGOTIATE:       negotiatedRender,
		encoding.NOOP:   noopRender,
		encoding.STRING: stringRender,
		encoding.JSON:   jsonRender,
	}
)

// RegisterRender allows clients to register their custom renders
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
		xmlRender(c, response)
	case gin.MIMEPlain:
		yamlRender(c, response)
	default:
		jsonRender(c, response)
	}
}

func stringRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.String(http.StatusOK, "")
		return
	}
	d, ok := response.Data["content"]
	if !ok {
		c.String(http.StatusOK, "")
		return
	}
	msg, ok := d.(string)
	if !ok {
		c.String(http.StatusOK, "")
		return
	}
	c.String(http.StatusOK, msg)
}

func jsonRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.JSON(http.StatusOK, emptyResponse)
		return
	}
	c.JSON(http.StatusOK, response.Data)
}

func xmlRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.XML(http.StatusOK, nil)
		return
	}
	d, ok := response.Data["content"]
	if !ok {
		c.XML(http.StatusOK, nil)
		return
	}
	c.XML(http.StatusOK, d)
}

func noopRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(response.Metadata.StatusCode)
	if hs, ok := response.Metadata.Headers["Content-Type"]; ok && len(hs) > 0 {
		c.Header("Content-Type", hs[0])
	}
	io.Copy(c.Writer, response.Io)
}

func yamlRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.YAML(http.StatusOK, emptyResponse)
		return
	}
	c.YAML(http.StatusOK, response.Data)
}

var emptyResponse = gin.H{}
