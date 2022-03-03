// SPDX-License-Identifier: Apache-2.0

package proxy

import "testing"

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
