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

func TestNewEngine_paramsAreChecked(t *testing.T) {
	engine := NewEngine(
		config.ServiceConfig{},
		EngineOptions{},
	)

	engine.GET("/user/:id/public", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	assertResponse := func(path string, statusCode int, body string) {
		req, _ := http.NewRequest("GET", path, http.NoBody)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		resp := w.Result()

		if sc := resp.StatusCode; sc != statusCode {
			t.Errorf("unexpected status code: %d (expected %d)", sc, statusCode)
			return
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("reading the response body: %s", err.Error())
			return
		}

		if string(b) != body {
			t.Errorf("unexpected response body: '%s' (expected '%s')", string(b), body)
		}
	}

	assertResponse("/user/123/public", http.StatusOK, "ok")
	assertResponse("/user/123%3f/public", http.StatusBadRequest, "error: encoded url params")
	assertResponse("/user/123%23/public", http.StatusBadRequest, "error: encoded url params")
}
