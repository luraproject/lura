package encoding

import (
	"strings"
	"testing"
)

func TestNewRawDecoder(t *testing.T) {
	decoder := NewRawDecoder(false)
	message := "somewhere over the rainbow"
	original := strings.NewReader(message)
	var result map[string]interface{}
	if err := decoder(original, &result); err != nil {
		t.Error("Unexpected error:", err.Error())
	}
	if len(result) != 1 {
		t.Error("Unexpected result:", result)
	}
	v, ok := result["content"]
	if !ok {
		t.Error("wrong result: content not found", result)
	}
	if v.(string) != message {
		t.Error("wrong result:", v)
	}
}
