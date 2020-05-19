package encoding

import (
	"io"
	"strings"
	"testing"
)

func BenchmarkDecoder(b *testing.B) {
	for _, dec := range []decoderTestCase{
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
				benchmarkDecoder(b, tc.input, dec.decoder)
			})
		}
	}
}

func benchmarkDecoder(b *testing.B, input string, dec func(io.Reader, *map[string]interface{}) error) {
	var result map[string]interface{}
	for i := 0; i < b.N; i++ {
		_ = dec(strings.NewReader(`["foo", "bar", "supu"]`), &result)
	}
}

type decoderTestCase struct {
	name    string
	decoder func(io.Reader, *map[string]interface{}) error
}
