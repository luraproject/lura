package proxy

import (
	"context"
	"testing"
)

func BenchmarkNewLoadBalancedMiddleware(b *testing.B) {
	proxy := newLoadBalancedMiddleware(dummyBalancer("supu"))(dummyProxy(&Response{}))
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		proxy(context.Background(), &Request{
			Path: "/tupu",
		})
	}
}
