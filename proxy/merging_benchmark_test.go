package proxy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
)

func BenchmarkNewMergeDataMiddleware(b *testing.B) {
	backend := config.Backend{}
	backends := []*config.Backend{&backend, &backend, &backend, &backend}

	partial1 := dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true})
	partial2 := dummyProxy(&Response{Data: map[string]interface{}{"tupu": true}, IsComplete: true})
	partial3 := dummyProxy(&Response{Data: map[string]interface{}{"foo": "bar"}, IsComplete: true})
	partial4 := dummyProxy(&Response{Data: map[string]interface{}{"foobar": false}, IsComplete: true})
	proxies := []Proxy{partial1, partial2, partial3, partial4}

	for testCase, totalParts := range []int{2, 3, 4} {
		b.Run(fmt.Sprintf("with %d parts", totalParts), func(b *testing.B) {
			endpoint := config.EndpointConfig{
				Backend: backends[:testCase+2],
				Timeout: time.Duration(100) * time.Millisecond,
			}
			proxy := NewMergeDataMiddleware(&endpoint)(proxies[:testCase+2]...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				proxy(context.Background(), &Request{})
			}
		})
	}
}
