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
	backends := make([]*config.Backend, 10)
	for i := range backends {
		backends[i] = &backend
	}

	proxies := []Proxy{
		dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"tupu": true}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"foo": "bar"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"foobar": false}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"qux": "false"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"data": "the quick brow fox"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"status": "ok"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"aaaa": "aaaaaaaaaaaa"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"bbbbb": 3.14}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"cccc": map[string]interface{}{"a": 42}}, IsComplete: true}),
	}

	for _, totalParts := range []int{2, 3, 4, 5, 6, 7, 8, 9, 10} {
		b.Run(fmt.Sprintf("with %d parts", totalParts), func(b *testing.B) {
			endpoint := config.EndpointConfig{
				Backend: backends[:totalParts],
				Timeout: time.Duration(100) * time.Millisecond,
			}
			proxy := NewMergeDataMiddleware(&endpoint)(proxies[:totalParts]...)
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				proxy(context.Background(), &Request{})
			}
		})
	}
}
