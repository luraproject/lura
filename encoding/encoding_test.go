package encoding

import (
	"strings"
	"testing"

	"github.com/devopsfaith/krakend/register"
)

func TestRegister(t *testing.T) {
	original := GetRegister()

	if len(original.data.Clone()) != 2 {
		t.Error("Unexpected number of registered factories:", len(original.data.Clone()))
	}

	decoders = &DecoderRegister{register.NewUntyped()}
	Register("some", NewJSONDecoder)

	if len(decoders.data.Clone()) != 1 {
		t.Error("Unexpected number of registered factories:", len(decoders.data.Clone()))
	}

	decoders = initDecoderRegister()
}

func TestGet(t *testing.T) {
	if len(decoders.data.Clone()) != 2 {
		t.Error("Unexpected number of registered factories:", len(decoders.data.Clone()))
	}

	checkDecoder(t, JSON)
	checkDecoder(t, "some")

	decoders = &DecoderRegister{register.NewUntyped()}
	Register("some", NewJSONDecoder)

	if len(decoders.data.Clone()) != 1 {
		t.Error("Unexpected number of registered factories:", len(decoders.data.Clone()))
	}

	checkDecoder(t, JSON)
	checkDecoder(t, "some")

	decoders = initDecoderRegister()
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
