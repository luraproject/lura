package gin

import (
	"io"
	"net/http"
	"sync"

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
	mutex          = &sync.RWMutex{}
	renderRegister = map[string]Render{
		NEGOTIATE:       negotiatedRender,
		encoding.STRING: stringRender,
		encoding.JSON:   jsonRender,
		encoding.NOOP:   noopRender,
	}
)

// RegisterRender allows clients to register their custom renders
func RegisterRender(name string, r Render) {
	mutex.Lock()
	renderRegister[name] = r
	mutex.Unlock()
}

func getRender(cfg *config.EndpointConfig) Render {
	fallback := jsonRender
	if len(cfg.Backend) == 1 {
		fallback = getWithFallback(cfg.Backend[0].Encoding, fallback)
	}

	if cfg.OutputEncoding == "" {
		return fallback
	}

	return getWithFallback(cfg.OutputEncoding, fallback)
}

func getWithFallback(key string, fallback Render) Render {
	mutex.RLock()
	r, ok := renderRegister[key]
	mutex.RUnlock()
	if !ok {
		return fallback
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

func yamlRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.YAML(http.StatusOK, emptyResponse)
		return
	}
	c.YAML(http.StatusOK, response.Data)
}

func noopRender(c *gin.Context, response *proxy.Response) {
	if response == nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(response.Metadata.StatusCode)
	for k, v := range response.Metadata.Headers {
		c.Header(k, v[0])
	}
	io.Copy(c.Writer, response.Io)
}

var emptyResponse = gin.H{}
