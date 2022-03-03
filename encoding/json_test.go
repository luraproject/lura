// SPDX-License-Identifier: Apache-2.0

package encoding

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func ExampleNewJSONDecoder_map() {
	decoder := NewJSONDecoder(false)
	original := strings.NewReader(`{"foo": "bar", "supu": false, "tupu": 4.20}`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		fmt.Println("Unexpected error:", err.Error())
	}
	fmt.Printf("%+v\n", result)

	// output:
	// map[foo:bar supu:false tupu:4.20]
}

func ExampleNewJSONDecoder_collection() {
	decoder := NewJSONDecoder(true)
	original := strings.NewReader(`["foo", "bar", "supu"]`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		fmt.Println("Unexpected error:", err.Error())
	}
	fmt.Printf("%+v\n", result)

	// output:
	// map[collection:[foo bar supu]]
}

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

func ExampleNewSafeJSONDecoder() {
	decoder := NewSafeJSONDecoder(true)
	for _, body := range []string{
		`{"foo": "bar", "supu": false, "tupu": 4.20}`,
		`["foo", "bar", "supu"]`,
	} {
		var result map[string]interface{}
		if err := decoder(strings.NewReader(body), &result); err != nil {
			fmt.Println("Unexpected error:", err.Error())
		}
		fmt.Printf("%+v\n", result)
	}

	// output:
	// map[foo:bar supu:false tupu:4.20]
	// map[collection:[foo bar supu]]
}

func TestNewSafeJSONDecoder_map(t *testing.T) {
	decoder := NewSafeJSONDecoder(false)
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

func TestNewSafeJSONDecoder_collection(t *testing.T) {
	decoder := NewSafeJSONDecoder(true)
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

func TestNewSafeJSONDecoder_other(t *testing.T) {
	decoder := NewSafeJSONDecoder(true)
	original := strings.NewReader(`3`)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if v, ok := result["result"]; !ok || v.(json.Number).String() != "3" {
		t.Error("wrong result:", result)
	}
}
