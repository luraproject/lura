package gin

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luraproject/lura/v2/config"
)

func TestNewEngine_contextIsPropagated(t *testing.T) {
	engine := NewEngine(
		config.ServiceConfig{},
		EngineOptions{},
	)

	type ctxKeyType string

	ctxKey := ctxKeyType("foo")
	ctxValue := "bar"

	engine.GET("/some/path", func(c *gin.Context) {
		c.String(http.StatusOK, "%v", c.Value(ctxKey))
	})

	req, _ := http.NewRequest("GET", "/some/path", http.NoBody)
	req = req.WithContext(context.WithValue(req.Context(), ctxKey, ctxValue))

	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	resp := w.Result()

	if sc := resp.StatusCode; sc != http.StatusOK {
		t.Errorf("unexpected status code: %d", sc)
		return
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("reading the response body: %s", err.Error())
		return
	}

	if string(b) != ctxValue {
		t.Errorf("unexpected value: %s", string(b))
	}
}
