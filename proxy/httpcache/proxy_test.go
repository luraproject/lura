package httpcache

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gregjones/httpcache"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
	"github.com/devopsfaith/krakend/proxy"
)

func TestHTTPProxy_ok(t *testing.T) {
	expectedMethod := "GET"
	counter := 0
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectedMethod {
			t.Errorf("Wrong request method. Want: %s. Have: %s", expectedMethod, r.Method)
		}
		counter++
		w.Header().Set("Cache-Control", "public, max-age=10")
		fmt.Fprintf(w, "{\"supu\":42, \"tupu\":true, \"foo\": \"bar\"}")
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{
		Decoder: encoding.JSONDecoder,
	}
	request := proxy.Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	totalRequests := 10
	p := NewHTTPProxy(httpcache.NewMemoryCacheTransport())(&backend)
	for i := 0; i < totalRequests; i++ {
		mustEnd := time.After(time.Duration(150) * time.Millisecond)

		result, err := p(context.Background(), &request)
		if err != nil {
			t.Errorf("The proxy returned an unexpected error: %s\n", err.Error())
			return
		}
		if result == nil {
			t.Errorf("The proxy returned a null result\n")
			return
		}
		select {
		case <-mustEnd:
			t.Errorf("Error: expected response")
			return
		default:
		}

		tmp, ok := result.Data["supu"]
		if !ok {
			t.Errorf("The proxy returned an unexpected result: %v\n", result)
		}
		supuValue, err := tmp.(json.Number).Int64()
		if err != nil || supuValue != 42 {
			t.Errorf("The proxy returned an unexpected result: %v\n", supuValue)
		}
		if v, ok := result.Data["tupu"]; !ok || !v.(bool) {
			t.Errorf("The proxy returned an unexpected result: %v\n", result)
		}
		if v, ok := result.Data["foo"]; !ok || v.(string) != "bar" {
			t.Errorf("The proxy returned an unexpected result: %v\n", result)
		}
	}
	if counter != 1 {
		t.Errorf("The counter returned an unexpected count: %d\n", counter)
	}
}

func newDummyReadCloser(content string) io.ReadCloser {
	return dummyReadCloser{strings.NewReader(content)}
}

type dummyReadCloser struct {
	reader io.Reader
}

func (d dummyReadCloser) Read(p []byte) (int, error) {
	return d.reader.Read(p)
}

func (d dummyReadCloser) Close() error {
	return nil
}
