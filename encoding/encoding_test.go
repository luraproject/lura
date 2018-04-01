package encoding

import (
	"strings"
	"testing"
)

func TestRegister(t *testing.T) {
	original := decoders

	if len(decoders) != 2 {
		t.Error("Unexpected number of registered factories:", len(decoders))
	}

	decoders = map[string]DecoderFactory{}
	Register("some", NewJSONDecoder)

	if len(decoders) != 1 {
		t.Error("Unexpected number of registered factories:", len(decoders))
	}

	decoders = original
}

func TestGet(t *testing.T) {
	original := decoders

	if len(decoders) != 2 {
		t.Error("Unexpected number of registered factories:", len(decoders))
	}

	checkDecoder(t, JSON)
	checkDecoder(t, "some")

	decoders = map[string]DecoderFactory{}
	Register("some", NewJSONDecoder)

	if len(decoders) != 1 {
		t.Error("Unexpected number of registered factories:", len(decoders))
	}

	checkDecoder(t, JSON)
	checkDecoder(t, "some")

	decoders = original
}

func checkDecoder(t *testing.T, name string) {
	d := Get(name)(false)

	input := strings.NewReader(`{"foo": "bar"}`)
	var result map[string]interface{}
	if err := d(input, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if result["foo"] != "bar" {
		t.Error("Unexpected value:", result["foo"])
	}
}
