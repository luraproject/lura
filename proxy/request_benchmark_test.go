// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"io"
	"strings"
	"testing"
)

func BenchmarkRequestGeneratePath(b *testing.B) {
	r := Request{
		Method: "GET",
		Params: map[string]string{
			"Supu": "42",
			"Tupu": "false",
			"Foo":  "bar",
		},
	}

	for _, testCase := range []string{
		"/a",
		"/a/{{.Supu}}",
		"/a?b={{.Tupu}}",
		"/a/{{.Supu}}/foo/{{.Foo}}",
		"/a/{{.Supu}}/foo/{{.Foo}}/b?c={{.Tupu}}",
	} {
		b.Run(testCase, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.GeneratePath(testCase)
			}
		})
	}
}

func BenchmarkCloneRequest(b *testing.B) {
	body := `{"id":1,"name":"test","items":[1,2,3],"nested":{"key":"value"}}`

	b.Run("with_body", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			clone := CloneRequest(newBenchmarkCloneRequestWithBody(body))
			_ = clone
		}
	})

	b.Run("with_body_parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				clone := CloneRequest(newBenchmarkCloneRequestWithBody(body))
				_ = clone
			}
		})
	})

	b.Run("without_body", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			clone := CloneRequest(newBenchmarkCloneRequestWithoutBody())
			_ = clone
		}
	})
}

func newBenchmarkCloneRequestWithBody(body string) *Request {
	return &Request{
		Method: "POST",
		Params: map[string]string{
			"Supu": "42",
			"Tupu": "false",
			"Foo":  "bar",
		},
		Headers: map[string][]string{
			"Content-Type":    {"application/json"},
			"Accept":          {"application/json"},
			"X-Forwarded-For": {"10.0.0.1"},
		},
		Body: io.NopCloser(strings.NewReader(body)),
	}
}

func newBenchmarkCloneRequestWithoutBody() *Request {
	return &Request{
		Method: "GET",
		Params: map[string]string{
			"Supu": "42",
			"Tupu": "false",
			"Foo":  "bar",
		},
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
			"Accept":       {"application/json"},
		},
	}
}
