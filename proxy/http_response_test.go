package proxy

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultHTTPResponseParserConfig_nopDecoder(t *testing.T) {
	result := map[string]interface{}{}
	if err := DefaultHTTPResponseParserConfig.Decoder(bytes.NewBufferString("some body"), &result); err != nil {
		t.Error(err.Error())
	}
	if len(result) != 0 {
		t.Error("unexpected result")
	}
}

func TestDefaultHTTPResponseParserConfig_nopEntityFormatter(t *testing.T) {
	expected := Response{Data: map[string]interface{}{"supu": "tupu"}, IsComplete: true}
	result := DefaultHTTPResponseParserConfig.EntityFormatter.Format(expected)
	if !result.IsComplete {
		t.Error("unexpected result")
	}
	d, ok := result.Data["supu"]
	if !ok {
		t.Error("unexpected result")
	}
	if v, ok := d.(string); !ok || v != "tupu" {
		t.Error("unexpected result")
	}
}

func TestNopHTTPResponseParser(t *testing.T) {
	w := httptest.NewRecorder()
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("header1", "value1")
		w.Write([]byte("some nice, interesting and long content"))
	}
	req, _ := http.NewRequest("GET", "/url", nil)
	handler(w, req)
	result, err := NoOpHTTPResponseParser(context.Background(), w.Result())
	if !result.IsComplete {
		t.Error("unexpected result")
	}
	if len(result.Data) != 0 {
		t.Error("unexpected result")
	}
	if result.Metadata.StatusCode != http.StatusOK {
		t.Error("unexpected result")
	}
	headers := result.Metadata.Headers
	if h, ok := headers["Header1"]; !ok || h[0] != "value1" {
		t.Error("unexpected result:", result.Metadata.Headers)
	}
	body, err := ioutil.ReadAll(result.Io)
	if err != nil {
		t.Error("unexpected error:", err.Error())
	}
	if string(body) != "some nice, interesting and long content" {
		t.Error("unexpected result")
	}
}
