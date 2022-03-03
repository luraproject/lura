// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"testing"

	"github.com/luraproject/lura/v2/config"
)

func BenchmarkNewRequestBuilderMiddleware(b *testing.B) {
	backend := config.Backend{
		URLPattern: "/supu",
		Method:     "GET",
	}
	proxy := NewRequestBuilderMiddleware(&backend)(dummyProxy(&Response{}))

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		proxy(context.Background(), &Request{})
	}
}
