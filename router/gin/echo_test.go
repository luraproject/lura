// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEchoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/_gin_endpoint/:param", EchoHandler())

	reqBody := `{"message":"some body to send"}`
	req := httptest.NewRequest("GET", "http://127.0.0.1:8088/_gin_endpoint/a?b=1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body, ioerr := io.ReadAll(w.Result().Body)
	if ioerr != nil {
		t.Error("reading a response:", ioerr.Error())
		return
	}
	w.Result().Body.Close()

	expectedBody := `{"Body":"{\"message\":\"some body to send\"}","Headers":{"Content-Type":["application/json"]},"Method":"GET","Params":[{"Key":"param","Value":"a"}],"Query":{"b":["1"]},"URL":"http://127.0.0.1:8088/_gin_endpoint/a?b=1"}`
	content := string(body)
	if w.Result().Header.Get("Cache-Control") != "" {
		t.Error("Cache-Control error:", w.Result().Header.Get("Cache-Control"))
	}
	if w.Result().Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Error("Content-Type error:", w.Result().Header.Get("Content-Type"))
	}
	if w.Result().Header.Get("X-Krakend") != "" {
		t.Error("X-Krakend error:", w.Result().Header.Get("X-Krakend"))
	}
	if w.Result().StatusCode != http.StatusOK {
		t.Error("Unexpected status code:", w.Result().StatusCode)
	}
	if content != expectedBody {
		t.Error("Unexpected body:", content, "expected:", expectedBody)
	}
}
