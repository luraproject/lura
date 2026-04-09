// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"encoding/json"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// legacyNew reproduces the original New() behavior before the fix, for comparison.
func legacyNew(opt Options) *Extractor {
	var replacements [][2]string

	title := cases.Title(language.Und)
	for k, v := range opt.Variables {
		val, ok := v.(string)
		if !ok {
			continue
		}
		if len(val) > 0 && val[0] == '{' && val[len(val)-1] == '}' {
			replacements = append(replacements, [2]string{k, title.String(val[1:2]) + val[2:len(val)-1]})
		}
	}

	paramExtractor := func(params map[string]string) (*GraphQLRequest, error) {
		val := GraphQLRequest{
			Query:         opt.Query,
			OperationName: opt.OperationName,
			Variables:     map[string]interface{}{},
		}
		for k, v := range opt.Variables {
			val.Variables[k] = v
		}
		for _, vs := range replacements {
			val.Variables[vs[0]] = params[vs[1]]
		}
		return &val, nil
	}

	return &Extractor{
		cfg:            opt,
		paramExtractor: paramExtractor,
		newBody: func(params map[string]string) ([]byte, error) {
			val, err := paramExtractor(params)
			if err != nil {
				return []byte{}, err
			}
			return json.Marshal(val)
		},
	}
}

var benchCfg = config.ExtraConfig{
	Namespace: map[string]interface{}{
		"type":  OperationQuery,
		"query": "{ search(query: $query) { nodes { id } } }",
		"variables": map[string]interface{}{
			"single":   "{owner}",
			"compound": "repo:{owner}/{repo} category:Announcements",
			"static":   "no-params-here",
		},
	},
}

var benchParams = map[string]string{
	"Owner": "krakend",
	"Repo":  "lura",
}

func BenchmarkBodyFromParams_legacy(b *testing.B) {
	cfg, _ := GetOptions(benchCfg)
	extractor := legacyNew(*cfg)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.BodyFromParams(benchParams)
	}
}

func BenchmarkBodyFromParams_fixed(b *testing.B) {
	cfg, _ := GetOptions(benchCfg)
	extractor := New(*cfg)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractor.BodyFromParams(benchParams)
	}
}
