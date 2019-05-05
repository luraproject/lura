package encoding

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"strings"
	"testing"
)

func encode(v interface{}) (*bytes.Buffer, error) {
	original := new(bytes.Buffer)
	err := gob.NewEncoder(original).Encode(v)
	return original, err
}

func TestNewGobDecoder_map(t *testing.T) {
	in := map[string]interface{}{
		"foo":  "bar",
		"supu": false,
		"tupu": 4.20,
	}
	original, err := encode(in)
	if err != nil {
		t.Errorf("encode: %s", err.Error())
		return
	}

	decoder := NewGobDecoder(false)

	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if !reflect.DeepEqual(result, in) {
		t.Errorf("unexpected response")
	}
	if len(result) != 3 {
		t.Error("Unexpected result:", result)
	}
	if v, ok := result["foo"]; !ok || v.(string) != "bar" {
		t.Error("wrong result:", result)
	}
	if v, ok := result["supu"]; !ok || v.(bool) {
		t.Error("wrong result:", result)
	}
	if v, ok := result["tupu"]; !ok || v.(float64) != 4.20 {
		t.Error("wrong result:", result)
	}
}

func TestNewGobDecoder_collection(t *testing.T) {
	in := []interface{}{"foo", "bar", "supu"}
	original, err := encode(in)
	if err != nil {
		t.Errorf("encode: %s", err.Error())
		return
	}

	decoder := NewGobDecoder(true)

	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if len(result) != 1 {
		t.Error("Unexpected result:", result)
	}
	v, ok := result["collection"]
	if !ok {
		t.Error("wrong result:", result)
	}
	embedded := v.([]interface{})
	if embedded[0].(string) != "foo" {
		t.Error("wrong result:", result)
	}
	if embedded[1].(string) != "bar" {
		t.Error("wrong result:", result)
	}
	if embedded[2].(string) != "supu" {
		t.Error("wrong result:", result)
	}
}

func TestNewGobDecoder_ko(t *testing.T) {
	decoder := NewGobDecoder(true)
	original := strings.NewReader(`3`)
	var result map[string]interface{}
	if err := decoder(original, &result); err == nil {
		t.Error("Expecting error!")
	}
}
