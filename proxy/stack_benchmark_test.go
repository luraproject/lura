// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

var result interface{}

func BenchmarkProxyStack_single(b *testing.B) {
	backend := &config.Backend{
		ConcurrentCalls: 3,
		Timeout:         time.Duration(100) * time.Millisecond,
		Host:            []string{"supu:8080"},
		Method:          "GET",
		URLPattern:      "/a/{{.Tupu}}",
		DenyList:        []string{"map.aaaa"},
		Mapping:         map[string]string{"supu": "SUPUUUUU"},
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

	ef := NewEntityFormatter(backend)
	p := func(_ context.Context, _ *Request) (*Response, error) {
		res := ef.Format(Response{
			Data: map[string]interface{}{
				"supu": 42,
				"tupu": true,
				"foo":  "bar",
				"map":  map[string]interface{}{"aaaa": false},
				"col": []interface{}{
					map[string]interface{}{
						"a": 1,
						"b": 2,
					},
				},
			},
			IsComplete: true,
		})
		return &res, nil
	}
	p = NewRoundRobinLoadBalancedMiddleware(backend)(p)
	p = NewConcurrentMiddleware(backend)(p)
	p = NewRequestBuilderMiddleware(backend)(p)
	p = NewStaticMiddleware(logging.NoOp, cfg)(p)

	request := &Request{
		Method:  "GET",
		Body:    newDummyReadCloser(""),
		Params:  map[string]string{"Tupu": "true"},
		Headers: map[string][]string{},
	}

	var r *Response
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ = p(context.Background(), request)
	}
	result = r
}

func BenchmarkProxyStack_multi(b *testing.B) {
	backend := &config.Backend{
		ConcurrentCalls: 3,
		Timeout:         time.Duration(100) * time.Millisecond,
		Host:            []string{"supu:8080"},
		Method:          "GET",
		URLPattern:      "/a/{{.Tupu}}",
		DenyList:        []string{"map.aaaa"},
		Mapping:         map[string]string{"supu": "SUPUUUUU"},
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
				ef := NewEntityFormatter(backend)
				backendProxy[i] = func(_ context.Context, _ *Request) (*Response, error) {
					res := ef.Format(Response{
						Data: map[string]interface{}{
							"supu": 42,
							"tupu": true,
							"foo":  "bar",
							"map":  map[string]interface{}{"aaaa": false},
							"col": []interface{}{
								map[string]interface{}{
									"a": 1,
									"b": 2,
								},
							},
						},
						IsComplete: true,
					})
					return &res, nil
				}
				backendProxy[i] = NewRoundRobinLoadBalancedMiddleware(backend)(backendProxy[i])
				backendProxy[i] = NewConcurrentMiddleware(backend)(backendProxy[i])
				backendProxy[i] = NewRequestBuilderMiddleware(backend)(backendProxy[i])
			}
			p := NewMergeDataMiddleware(logging.NoOp, cfg)(backendProxy...)
			p = NewStaticMiddleware(logging.NoOp, cfg)(p)

			var r *Response
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r, _ = p(context.Background(), request)
			}
			result = r
		})
	}
}

func BenchmarkProxyStack_single_flatmap(b *testing.B) {
	backend := &config.Backend{
		ConcurrentCalls: 3,
		Timeout:         time.Duration(100) * time.Millisecond,
		Host:            []string{"supu:8080"},
		Method:          "GET",
		URLPattern:      "/a/{{.Tupu}}",
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				flatmapKey: []interface{}{
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"map.aaaa"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"supu", "SUPUUUUU"},
					},
				},
			},
		},
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

	ef := NewEntityFormatter(backend)
	p := func(_ context.Context, _ *Request) (*Response, error) {
		res := ef.Format(Response{
			Data: map[string]interface{}{
				"supu": 42,
				"tupu": true,
				"foo":  "bar",
				"map":  map[string]interface{}{"aaaa": false},
				"col": []interface{}{
					map[string]interface{}{
						"a": 1,
						"b": 2,
					},
				},
			},
			IsComplete: true,
		})
		return &res, nil
	}
	p = NewRoundRobinLoadBalancedMiddleware(backend)(p)
	p = NewConcurrentMiddleware(backend)(p)
	p = NewRequestBuilderMiddleware(backend)(p)
	p = NewStaticMiddleware(logging.NoOp, cfg)(p)

	request := &Request{
		Method:  "GET",
		Body:    newDummyReadCloser(""),
		Params:  map[string]string{"Tupu": "true"},
		Headers: map[string][]string{},
	}

	var r *Response
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r, _ = p(context.Background(), request)
	}
	result = r
}

func BenchmarkProxyStack_multi_flatmap(b *testing.B) {
	backend := &config.Backend{
		ConcurrentCalls: 3,
		Timeout:         time.Duration(100) * time.Millisecond,
		Host:            []string{"supu:8080"},
		Method:          "GET",
		URLPattern:      "/a/{{.Tupu}}",
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				flatmapKey: []interface{}{
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"map.aaaa"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"supu", "SUPUUUUU"},
					},
				},
			},
		},
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
				ef := NewEntityFormatter(backend)
				backendProxy[i] = func(_ context.Context, _ *Request) (*Response, error) {
					res := ef.Format(Response{
						Data: map[string]interface{}{
							"supu": 42,
							"tupu": true,
							"foo":  "bar",
							"map":  map[string]interface{}{"aaaa": false},
							"col": []interface{}{
								map[string]interface{}{
									"a": 1,
									"b": 2,
								},
							},
						},
						IsComplete: true,
					})
					return &res, nil
				}
				backendProxy[i] = NewRoundRobinLoadBalancedMiddleware(backend)(backendProxy[i])
				backendProxy[i] = NewConcurrentMiddleware(backend)(backendProxy[i])
				backendProxy[i] = NewRequestBuilderMiddleware(backend)(backendProxy[i])
			}
			p := NewMergeDataMiddleware(logging.NoOp, cfg)(backendProxy...)
			p = NewStaticMiddleware(logging.NoOp, cfg)(p)

			var r *Response
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				p(context.Background(), request)
			}
			result = r
		})
	}
}
