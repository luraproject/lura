package gin

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devopsfaith/krakend/logging/gologging"
)

func TestDebugHandler(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, err := gologging.NewLogger("ERROR", buff, "pref")
	if err != nil {
		t.Error("building the logger:", err.Error())
		return
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/_gin_endpoint/:param", DebugHandler(logger))
	s := &http.Server{
		Addr:    ":8088",
		Handler: router,
	}
	defer s.Shutdown(context.Background())
	go s.ListenAndServe()

	time.Sleep(5 * time.Millisecond)

	req, _ := http.NewRequest("GET", "http://127.0.0.1:8088/_gin_endpoint/a?b=1", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error("sending a request:", err.Error())
		return
	}
	defer resp.Body.Close()

	body, ioerr := ioutil.ReadAll(resp.Body)
	if ioerr != nil {
		t.Error("reading a response:", err.Error())
		return
	}

	expectedBody := "{\"message\":\"pong\"}"

	content := string(body)
	if resp.Header.Get("Cache-Control") != "" {
		t.Error("Cache-Control error:", resp.Header.Get("Cache-Control"))
	}
	if resp.Header.Get("Content-Type") != "application/json; charset=utf-8" {
		t.Error("Content-Type error:", resp.Header.Get("Content-Type"))
	}
	if resp.Header.Get("X-Krakend") != "" {
		t.Error("X-Krakend error:", resp.Header.Get("X-Krakend"))
	}
	if resp.StatusCode != http.StatusOK {
		t.Error("Unexpected status code:", resp.StatusCode)
	}
	if content != expectedBody {
		t.Error("Unexpected body:", content, "expected:", expectedBody)
	}
}
