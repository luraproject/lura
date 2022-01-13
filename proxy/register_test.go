// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"testing"
)

func TestNewRegister_responseCombiner_ok(t *testing.T) {
	r := NewRegister()
	r.SetResponseCombiner("name1", func(total int, parts []*Response) *Response {
		if total < 0 || total >= len(parts) {
			return nil
		}
		return parts[total]
	})

	rc, ok := r.GetResponseCombiner("name1")
	if !ok {
		t.Error("expecting response combiner")
		return
	}

	result := rc(0, []*Response{{IsComplete: true, Data: map[string]interface{}{"a": 42}}})

	if result == nil {
		t.Error("expecting result")
		return
	}

	if !result.IsComplete {
		t.Error("expecting a complete result")
		return
	}

	if len(result.Data) != 1 {
		t.Error("unexpected result size:", len(result.Data))
		return
	}
}

func TestNewRegister_responseCombiner_fallbackIfErrored(t *testing.T) {
	r := NewRegister()

	r.data.Register("errored", true)

	rc, ok := r.GetResponseCombiner("errored")
	if !ok {
		t.Error("expecting response combiner")
		return
	}

	original := &Response{IsComplete: true, Data: map[string]interface{}{"a": 42}}

	result := rc(0, []*Response{original})

	if result != original {
		t.Error("unexpected result:", result)
		return
	}
}

func TestNewRegister_responseCombiner_fallbackIfUnknown(t *testing.T) {
	r := NewRegister()

	rc, ok := r.GetResponseCombiner("unknown")
	if ok {
		t.Error("the response combiner should not be found")
		return
	}

	original := &Response{IsComplete: true, Data: map[string]interface{}{"a": 42}}

	result := rc(0, []*Response{original})

	if result != original {
		t.Error("unexpected result:", result)
		return
	}
}
