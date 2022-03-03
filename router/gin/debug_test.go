// SPDX-License-Identifier: Apache-2.0

package gin

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/luraproject/lura/v2/logging"
)

func TestDebugHandler(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := logging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	router := gin.New()
	router.GET("/_gin_endpoint/:param", DebugHandler(logger))

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8088/_gin_endpoint/a?b=1", ioutil.NopCloser(&bytes.Buffer{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	body, ioerr := ioutil.ReadAll(w.Result().Body)
	if ioerr != nil {
		t.Error("reading a response:", err.Error())
		return
	}
	w.Result().Body.Close()

	expectedBody := "{\"message\":\"pong\"}"

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
