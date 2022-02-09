// SPDX-License-Identifier: Apache-2.0

package backoff

import (
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	next := 1
	for i := 0; i < 10; i++ {
		if v := int(ExponentialBackoff(i) / time.Second); v != next {
			t.Errorf("have: %d, want: %d", v, next)
		}
		next *= 2
	}
}

func TestLinearBackoff(t *testing.T) {
	for i := 0; i < 10; i++ {
		if v := int(LinearBackoff(i) / time.Second); v != i {
			t.Errorf("have: %d, want: %d", v, i)
		}
	}
}

func TestDefaultBackoff(t *testing.T) {
	for i := 0; i < 10; i++ {
		if v := int(DefaultBackoff(i) / time.Second); v != 1 {
			t.Errorf("have: %d, want: %d", v, 1)
		}
	}
}
