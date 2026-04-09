// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"strings"
	"testing"
)

// legacyInterpolate reproduces the original per-request substitution: direct map lookup.
func legacyInterpolate(replacements [][2]string, variables map[string]interface{}, params map[string]string) {
	for _, vs := range replacements {
		variables[vs[0]] = params[vs[1]]
	}
}

// fixedInterpolate is the new per-request substitution: strings.ReplaceAll loop (mirrors proxy.GeneratePath).
func fixedInterpolate(templates [][2]string, variables map[string]interface{}, params map[string]string) {
	for _, tmpl := range templates {
		buff := tmpl[1]
		for k, v := range params {
			buff = strings.ReplaceAll(buff, "{{."+k+"}}", v)
		}
		variables[tmpl[0]] = buff
	}
}

var (
	// legacy: replacements built as [varKey, capitalizedParamKey]
	legacyReplacements = [][2]string{
		{"single", "Owner"},
	}

	// fixed: templates built as [varKey, "{{.Owner}}/{{.Repo}} ..."]
	fixedTemplates = [][2]string{
		{"single", "{{.Owner}}"},
		{"compound", "repo:{{.Owner}}/{{.Repo}} category:Announcements"},
	}

	benchVariables = map[string]interface{}{
		"single":   "{owner}",
		"compound": "repo:{owner}/{repo} category:Announcements",
		"static":   "no-params-here",
	}

	benchParams = map[string]string{
		"Owner": "krakend",
		"Repo":  "lura",
	}
)

func BenchmarkInterpolation_legacy(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		vars := map[string]interface{}{}
		for k, v := range benchVariables {
			vars[k] = v
		}
		legacyInterpolate(legacyReplacements, vars, benchParams)
	}
}

func BenchmarkInterpolation_fixed(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		vars := map[string]interface{}{}
		for k, v := range benchVariables {
			vars[k] = v
		}
		fixedInterpolate(fixedTemplates, vars, benchParams)
	}
}
