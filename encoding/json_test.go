package encoding

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewJSONDecoder_map(t *testing.T) {
	decoder := NewJSONDecoder(false)
	original := strings.NewReader(`{"foo": "bar", "supu": false, "tupu": 4.20}`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
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
	if v, ok := result["tupu"]; !ok || v.(json.Number).String() != "4.20" {
		t.Error("wrong result:", result)
	}
}

func TestNewJSONDecoder_collection(t *testing.T) {
	decoder := NewJSONDecoder(true)
	original := strings.NewReader(`["foo", "bar", "supu"]`)
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

func TestNewJSONDecoder_ko(t *testing.T) {
	decoder := NewJSONDecoder(true)
	original := strings.NewReader(`3`)
	var result map[string]interface{}
	if err := decoder(original, &result); err == nil {
		t.Error("Expecting error!")
	}
}
