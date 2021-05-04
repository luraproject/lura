// SPDX-License-Identifier: Apache-2.0
package encoding

import (
	"strings"
	"testing"
)

func TestNewStringDecoder_ok(t *testing.T) {
	decoder := NewStringDecoder(false)
	message := "somewhere over the rainbow"
	original := strings.NewReader(message)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error: ", err.Error())
	}
	if len(result) != 1 {
		t.Error("Unexpected result: ", result)
	}
	v, ok := result["content"]
	if !ok {
		t.Error("Wrong result: content not found ", result)
	}
	if v.(string) != message {
		t.Error("Wrong result: ", v)
	}
}

func TestNewStringDecoder_ko(t *testing.T) {
	decoder := NewStringDecoder(false)
	errorMsg := erroredReader("some error")
	var result map[string]interface{}
	if err := decoder(errorMsg, &result); err == nil || err.Error() != errorMsg.Error() {
		t.Error("Unexpected error:", err)
	}
}

type erroredReader string

func (e erroredReader) Error() string {
	return string(e)
}

func (e erroredReader) Read(_ []byte) (n int, err error) {
	return 0, e
}
