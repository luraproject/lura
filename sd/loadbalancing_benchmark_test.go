// SPDX-License-Identifier: Apache-2.0

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

func BenchmarkLB(b *testing.B) {
	for _, tc := range []struct {
		name string
		f    func([]string) Balancer
	}{
		{name: "round_robin", f: func(hs []string) Balancer { return NewRoundRobinLB(FixedSubscriber(hs)) }},
		{name: "random", f: func(hs []string) Balancer { return NewRandomLB(FixedSubscriber(hs)) }},
	} {
		for _, testCase := range balancerTestsCases {
			b.Run(fmt.Sprintf("%s/%d", tc.name, len(testCase)), func(b *testing.B) {
				balancer := tc.f(testCase)
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					balancer.Host()
				}
			})
		}
	}
}

func BenchmarkLB_parallel(b *testing.B) {
	for _, tc := range []struct {
		name string
		f    func([]string) Balancer
	}{
		{name: "round_robin", f: func(hs []string) Balancer { return NewRoundRobinLB(FixedSubscriber(hs)) }},
		{name: "random", f: func(hs []string) Balancer { return NewRandomLB(FixedSubscriber(hs)) }},
	} {
		for _, testCase := range balancerTestsCases {
			b.Run(fmt.Sprintf("%s/%d", tc.name, len(testCase)), func(b *testing.B) {
				balancer := tc.f(testCase)
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						balancer.Host()
					}
				})
			})
		}
	}
}
