// SPDX-License-Identifier: Apache-2.0

package encoding

import (
	"io"
	"strings"
	"testing"
)

func BenchmarkDecoder(b *testing.B) {
	for _, dec := range []struct {
		name    string
		decoder func(io.Reader, *map[string]interface{}) error
	}{
		{
			name:    "json-collection",
			decoder: NewJSONDecoder(true),
		},
		{
			name:    "json-map",
			decoder: NewJSONDecoder(false),
		},
		{
			name:    "safe-json-collection",
			decoder: NewSafeJSONDecoder(true),
		},
		{
			name:    "safe-json-map",
			decoder: NewSafeJSONDecoder(true),
		},
	} {
		for _, tc := range []struct {
			name  string
			input string
		}{
			{
				name:  "collection",
				input: `["a","b","c"]`,
			},
			{
				name:  "map",
				input: `{"foo": "bar", "supu": false, "tupu": 4.20}`,
			},
		} {
			b.Run(dec.name+"/"+tc.name, func(b *testing.B) {
				var result map[string]interface{}
				for i := 0; i < b.N; i++ {
					_ = dec.decoder(strings.NewReader(tc.input), &result)
				}
			})
		}
	}
}
