// SPDX-License-Identifier: Apache-2.0

package register

import "testing"

func TestNamespaced(t *testing.T) {
	r := New()
	r.Register("namespace1", "name1", 42)
	r.AddNamespace("namespace1")
	r.AddNamespace("namespace2")
	r.Register("namespace2", "name2", true)

	nr, ok := r.Get("namespace1")
	if !ok {
		t.Error("namespace1 not found")
		return
	}
	if _, ok := nr.Get("name2"); ok {
		t.Error("name2 found into namespace1")
		return
	}
	v1, ok := nr.Get("name1")
	if !ok {
		t.Error("name1 not found")
		return
	}
	if i, ok := v1.(int); !ok || i != 42 {
		t.Error("unexpected value:", v1)
	}

	nr, ok = r.Get("namespace2")
	if !ok {
		t.Error("namespace2 not found")
		return
	}
	if _, ok := nr.Get("name1"); ok {
		t.Error("name1 found into namespace2")
		return
	}
	v2, ok := nr.Get("name2")
	if !ok {
		t.Error("name2 not found")
		return
	}
	if b, ok := v2.(bool); !ok || !b {
		t.Error("unexpected value:", v2)
	}
}
