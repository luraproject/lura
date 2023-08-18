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
	reqBody := `{"message":"some body to send"}`
	expectedRespBody := `{"req_uri":"http://127.0.0.1:8088/_gin_endpoint/a?b=1","req_uri_details":{"fragment":"","host":"127.0.0.1:8088","path":"/_gin_endpoint/a","query":"b=1","user":""},"req_method":"GET","req_querystring":{"b":["1"]},"req_body":"{\"message\":\"some body to send\"}","req_headers":{"Content-Type":["application/json"]}}`
	expectedRespNoBody := `{"req_uri":"http://127.0.0.1:8088/_gin_endpoint/a?b=1","req_uri_details":{"fragment":"","host":"127.0.0.1:8088","path":"/_gin_endpoint/a","query":"b=1","user":""},"req_method":"GET","req_querystring":{"b":["1"]},"req_body":"","req_headers":{"Content-Type":["application/json"]}}`
	expectedRespString := `{"req_uri":"http://127.0.0.1:8088/_gin_endpoint/a?b=1","req_uri_details":{"fragment":"","host":"127.0.0.1:8088","path":"/_gin_endpoint/a","query":"b=1","user":""},"req_method":"GET","req_querystring":{"b":["1"]},"req_body":"Hello lura","req_headers":{"Content-Type":["application/json"]}}`

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/_gin_endpoint/:param", EchoHandler())

	for _, tc := range []struct {
		name string
		body io.Reader
		resp string
	}{
		{
			name: "json body",
			body: strings.NewReader(reqBody),
			resp: expectedRespBody,
		},
		{
			name: "no body",
			body: http.NoBody,
			resp: expectedRespNoBody,
		},
		{
			name: "string body",
			body: strings.NewReader("Hello lura"),
			resp: expectedRespString,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			echoRunTestRequest(t, router, tc.body, tc.resp)
		})
	}

}

func echoRunTestRequest(t *testing.T, e *gin.Engine, body io.Reader, expected string) {
	req := httptest.NewRequest("GET", "http://127.0.0.1:8088/_gin_endpoint/a?b=1", body)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	respBody, ioerr := io.ReadAll(w.Result().Body)
	if ioerr != nil {
		t.Error("reading a response:", ioerr.Error())
		return
	}
	w.Result().Body.Close()

	content := string(respBody)
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
	if content != expected {
		t.Error("Unexpected body:", content, "expected:", expected)
	}
}
