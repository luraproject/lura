package encoding

import (
	"bytes"
	"testing"
)

func TestRegister(t *testing.T) {
	original := decoders

	if len(decoders) != 3 {
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

	if len(decoders) != 3 {
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

func TestNoOpDecoder(t *testing.T) {
	d := Get(NOOP)(false)

	errorMsg := erroredReader("this error should never been sent")
	var result map[string]interface{}
	if err := d(errorMsg, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if result != nil {
		t.Error("Unexpected value:", result)
	}
}

func checkDecoder(t *testing.T, name string) {
	d := Get(name)(false)

	input := bytes.NewBufferString(`{"foo": "bar"}`)
	var result map[string]interface{}
	if err := d(input, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if result["foo"] != "bar" {
		t.Error("Unexpected value:", result["foo"])
	}
}
