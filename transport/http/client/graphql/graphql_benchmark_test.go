// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"testing"

	"github.com/luraproject/lura/v2/config"
)

func BenchmarkExtractor_BodyFromParams(b *testing.B) {
	params := map[string]string{
		"Owner": "krakend",
		"Repo":  "lura",
	}

	for _, tc := range []struct {
		name      string
		variables map[string]interface{}
	}{
		{
			name:      "no_params",
			variables: map[string]interface{}{"static": "no-params-here"},
		},
		{
			name:      "single_param",
			variables: map[string]interface{}{"owner": "{owner}"},
		},
		{
			name: "compound_params",
			variables: map[string]interface{}{
				"query": "repo:{owner}/{repo} category:Announcements",
			},
		},
		{
			name: "mixed",
			variables: map[string]interface{}{
				"single":   "{owner}",
				"compound": "repo:{owner}/{repo} category:Announcements",
				"static":   "no-params-here",
			},
		},
	} {
		cfg, _ := GetOptions(config.ExtraConfig{
			Namespace: map[string]interface{}{
				"type":      OperationQuery,
				"query":     "{ search(query: $query) { nodes { id } } }",
				"variables": tc.variables,
			},
		})
		extractor := New(*cfg)

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				extractor.BodyFromParams(params)
			}
		})
	}
}
