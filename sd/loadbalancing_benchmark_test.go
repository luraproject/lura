package sd

import (
	"fmt"
	"testing"
)

var balancerTestsCases = [][]string{
	{"a"},
	{"a", "b", "c"},
	{"a", "b", "c", "e", "f"},
}

func BenchmarkRoundRobinLB(b *testing.B) {
	for _, testCase := range balancerTestsCases {
		b.Run(fmt.Sprintf("%d hosts", len(testCase)), func(b *testing.B) {
			balancer := NewRoundRobinLB(FixedSubscriber(testCase))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				balancer.Host()
			}
		})
	}
}

func BenchmarkRoundRobinLB_parallel(b *testing.B) {
	for _, testCase := range balancerTestsCases {
		b.Run(fmt.Sprintf("%d hosts", len(testCase)), func(b *testing.B) {
			balancer := NewRoundRobinLB(FixedSubscriber(testCase))
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					balancer.Host()
				}
			})
		})
	}
}

func BenchmarkRandomLB(b *testing.B) {
	for _, testCase := range balancerTestsCases {
		b.Run(fmt.Sprintf("%d hosts", len(testCase)), func(b *testing.B) {
			balancer := NewRandomLB(FixedSubscriber(testCase), 1415926)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				balancer.Host()
			}
		})
	}
}
