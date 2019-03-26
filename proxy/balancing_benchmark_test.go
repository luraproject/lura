package proxy

import (
	"context"
	"testing"
)

func BenchmarkNewLoadBalancedMiddleware(b *testing.B) {
	for _, tc := range []struct {
		name string
		host string
		path string
	}{
		{name: "3", host: "abc", path: "abc"},
		{name: "5", host: "abcde", path: "abcde"},
		{name: "9", host: "abcdefghi", path: "abcdefghi"},
		{name: "13", host: "abcdefghijklm", path: "abcdefghijklm"},
		{name: "17", host: "abcdefghijklmopqr", path: "abcdefghijklmopqr"},
		{name: "21", host: "abcdefghijklmopqrstuv", path: "abcdefghijklmopqrstuv"},
		{name: "25", host: "abcdefghijklmopqrstuvwxyz", path: "abcdefghijklmopqrstuvwxyz"},
		{
			name: "50",
			host: "abcdefghijklmopqrstuvwxyzabcdefghijklmopqrstuvwxyz",
			path: "abcdefghijklmopqrstuvwxyzabcdefghijklmopqrstuvwxyz",
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			proxy := newLoadBalancedMiddleware(dummyBalancer(tc.host))(dummyProxy(&Response{}))
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				proxy(context.Background(), &Request{
					Path: tc.path,
				})
			}
		})
	}
}
