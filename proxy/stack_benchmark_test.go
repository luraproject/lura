package proxy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
)

func BenchmarkProxyStack_single(b *testing.B) {
	backend := &config.Backend{
		ConcurrentCalls: 3,
		Timeout:         time.Duration(100) * time.Millisecond,
		Host:            []string{"supu:8080"},
		Method:          "GET",
		URLPattern:      "/a/{{.Tupu}}",
	}
	cfg := &config.EndpointConfig{
		Backend: []*config.Backend{backend},
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				staticKey: map[string]interface{}{
					"data": map[string]interface{}{
						"status": "errored",
					},
				},
				"strategy": "incomplete",
			},
		},
	}
	expected := Response{
		Data:       map[string]interface{}{"supu": 42, "tupu": true, "foo": "bar"},
		IsComplete: true,
	}

	p := dummyProxy(&expected)
	p = NewRoundRobinLoadBalancedMiddleware(backend)(p)
	p = NewConcurrentMiddleware(backend)(p)
	p = NewRequestBuilderMiddleware(backend)(p)
	p = NewStaticMiddleware(cfg)(p)

	request := &Request{
		Method:  "GET",
		Body:    newDummyReadCloser(""),
		Params:  map[string]string{"Tupu": "true"},
		Headers: map[string][]string{},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p(context.Background(), request)
	}
}

func BenchmarkProxyStack_multi(b *testing.B) {
	backend := &config.Backend{
		ConcurrentCalls: 3,
		Timeout:         time.Duration(100) * time.Millisecond,
		Host:            []string{"supu:8080"},
		Method:          "GET",
		URLPattern:      "/a/{{.Tupu}}",
	}
	expected := Response{
		Data:       map[string]interface{}{"supu": 42, "tupu": true, "foo": "bar"},
		IsComplete: true,
	}

	request := &Request{
		Method:  "GET",
		Body:    newDummyReadCloser(""),
		Params:  map[string]string{"Tupu": "true"},
		Headers: map[string][]string{},
	}

	for _, testCase := range [][]*config.Backend{
		{backend},
		{backend, backend},
		{backend, backend, backend},
		{backend, backend, backend, backend},
		{backend, backend, backend, backend, backend},
	} {
		b.Run(fmt.Sprintf("with %d backends", len(testCase)), func(b *testing.B) {

			cfg := &config.EndpointConfig{
				Backend: testCase,
			}

			backendProxy := make([]Proxy, len(cfg.Backend))

			for i, backend := range cfg.Backend {
				backendProxy[i] = dummyProxy(&expected)
				backendProxy[i] = NewRoundRobinLoadBalancedMiddleware(backend)(backendProxy[i])
				backendProxy[i] = NewConcurrentMiddleware(backend)(backendProxy[i])
				backendProxy[i] = NewRequestBuilderMiddleware(backend)(backendProxy[i])
			}
			p := NewMergeDataMiddleware(cfg)(backendProxy...)
			p = NewStaticMiddleware(cfg)(p)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p(context.Background(), request)
			}
		})
	}
}
