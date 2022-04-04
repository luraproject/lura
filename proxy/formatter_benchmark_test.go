// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"github.com/luraproject/lura/v2/config"
)

func BenchmarkEntityFormatter_allowFilter(b *testing.B) {
	data := map[string]interface{}{
		"supu": 42,
		"tupu": false,
		"foo":  "bar",
	}

	for _, extraFields := range []int{0, 5, 10, 15, 20, 25} {
		sampleData := data
		for i := 0; i < extraFields; i++ {
			sampleData[fmt.Sprintf("%d", i)] = i
		}
		for _, testCase := range [][]string{
			{},
			{"supu"},
			{"supu", "tupu"},
			{"supu", "tupu", "foo"},
			{"supu", "tupu", "foo", "unknown"},
		} {
			sample := Response{
				Data:       sampleData,
				IsComplete: true,
			}
			b.Run(fmt.Sprintf("with %d elements with %d extra fields", len(testCase), extraFields), func(b *testing.B) {
				f := NewEntityFormatter(&config.Backend{AllowList: testCase})
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					f.Format(sample)
				}
			})
		}
	}

}

func benchmarkDeepChilds(depth, extraSiblings int) map[string]interface{} {
	data := make(map[string]interface{}, extraSiblings+1)
	for i := 0; i < extraSiblings; i++ {
		data[fmt.Sprintf("extra%d", i)] = "sibling_value"
	}
	if depth > 0 {
		data[fmt.Sprintf("child%d", depth)] = benchmarkDeepChilds(depth-1, extraSiblings)
	} else {
		data["child0"] = 1
	}
	return data
}

func benchmarkDeepStructure(numTargets, targetDepth, extraFields, extraSiblings int) (map[string]interface{}, []string) {
	data := make(map[string]interface{}, numTargets+extraFields)
	targetKeys := make([]string, numTargets)
	for i := 0; i < numTargets; i++ {
		data[fmt.Sprintf("target%d", i)] = benchmarkDeepChilds(targetDepth-1, extraSiblings)
	}
	for j := 0; j < extraFields; j++ {
		data[fmt.Sprintf("extra%d", j)] = benchmarkDeepChilds(targetDepth-1, extraSiblings)
	}
	// create the target list
	for i := 0; i < numTargets; i++ {
		var buffer bytes.Buffer
		buffer.WriteString(fmt.Sprintf("target%d", i))
		for j := targetDepth - 1; j >= 0; j-- {
			buffer.WriteString(fmt.Sprintf(".child%d", j))
		}
		targetKeys[i] = buffer.String()
	}
	return data, targetKeys
}

func BenchmarkEntityFormatter_deepAllowFilter(b *testing.B) {
	numTargets := []int{0, 1, 2, 5, 10}
	depths := []int{1, 3, 7}
	for _, nTargets := range numTargets {
		for _, depth := range depths {
			extraFields := nTargets + depth*2
			extraSiblings := nTargets
			data, allow := benchmarkDeepStructure(nTargets, depth, extraFields, extraSiblings)
			sample := Response{
				Data:       data,
				IsComplete: true,
			}
			f := NewEntityFormatter(&config.Backend{AllowList: allow})
			b.Run(fmt.Sprintf("numTargets:%d,depth:%d,extraFields:%d,extraSiblings:%d", nTargets, depth, extraFields, extraSiblings), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					f.Format(sample)
				}
			})
		}
	}
}

func BenchmarkEntityFormatter_denyFilter(b *testing.B) {
	data := map[string]interface{}{
		"supu": 42,
		"tupu": false,
		"foo":  "bar",
	}

	for _, extraFields := range []int{0, 5, 10, 15, 20, 25} {
		sampleData := data
		for i := 0; i < extraFields; i++ {
			sampleData[fmt.Sprintf("%d", i)] = i
		}
		for _, testCase := range [][]string{
			{},
			{"supu"},
			{"supu", "tupu"},
			{"supu", "tupu", "foo"},
			{"supu", "tupu", "foo", "unknown"},
		} {
			sample := Response{
				Data:       sampleData,
				IsComplete: true,
			}
			b.Run(fmt.Sprintf("with %d elements with %d extra fields", len(testCase), extraFields), func(b *testing.B) {
				f := NewEntityFormatter(&config.Backend{DenyList: testCase})
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					f.Format(sample)
				}
			})
		}
	}
}

func BenchmarkEntityFormatter_grouping(b *testing.B) {
	preffix := "group1"
	for _, extraFields := range []int{0, 5, 10, 15, 20, 25} {
		sampleData := make(map[string]interface{}, extraFields)
		for i := 0; i < extraFields; i++ {
			sampleData[fmt.Sprintf("%d", i)] = i
		}
		sample := Response{
			Data:       sampleData,
			IsComplete: true,
		}
		b.Run(fmt.Sprintf("with %d elements", extraFields), func(b *testing.B) {
			f := NewEntityFormatter(&config.Backend{Group: preffix})
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				f.Format(sample)
			}
		})
	}
}

func BenchmarkEntityFormatter_mapping(b *testing.B) {
	for _, extraFields := range []int{0, 5, 10, 15, 20, 25} {
		sampleData := make(map[string]interface{}, extraFields)
		for i := 0; i < extraFields; i++ {
			sampleData[fmt.Sprintf("%d", i)] = i
		}
		for _, testCase := range []map[string]string{
			{},
			{"1": "supu"},
			{"1": "supu", "2": "tupu"},
			{"1": "supu", "2": "tupu", "3": "foo"},
			{"1": "supu", "2": "tupu", "3": "foo", "4": "bar"},
			{"1": "supu", "2": "tupu", "3": "foo", "4": "bar", "5": "a"},
		} {
			sample := Response{
				Data:       sampleData,
				IsComplete: true,
			}
			b.Run(fmt.Sprintf("with %d elements with %d extra fields", len(testCase), extraFields), func(b *testing.B) {
				f := NewEntityFormatter(&config.Backend{Mapping: testCase})
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					f.Format(sample)
				}
			})
		}
	}
}

func BenchmarkEntityFormatter_flatmapAlt(b *testing.B) {
	f := NewEntityFormatter(&config.Backend{
		Target: "content",
		Group:  "group",
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				flatmapKey: []interface{}{
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"c"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"supu", "SUPUUUUU"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"a.b", "a.BOOOOO"},
					},
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"collection.*.b"},
					},
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"collection.*.d"},
					},
					map[string]interface{}{
						"type": "del",
						"args": []interface{}{"collection.*.e"},
					},
					map[string]interface{}{
						"type": "move",
						"args": []interface{}{"collection.*.c", "collection.*.x"},
					},
				},
			},
		},
	})

	for _, size := range []int{1, 2, 5, 10, 20, 50, 100, 500} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			sub := map[string]interface{}{
				"b": true,
				"c": 42,
				"d": "tupu",
				"e": []interface{}{1, 2, 3, 4},
			}
			sample := Response{
				Data: map[string]interface{}{
					"content": map[string]interface{}{
						"supu":       42,
						"tupu":       false,
						"foo":        "bar",
						"a":          sub,
						"collection": []interface{}{sub, sub, sub, sub},
					},
				},
				IsComplete: true,
			}
			subCol := []interface{}{}
			for i := 0; i < size; i++ {
				subCol = append(subCol, i)
			}
			sub["e"] = subCol
			sampleSubCol := []interface{}{}
			for i := 0; i < size; i++ {
				sampleSubCol = append(sampleSubCol, sub)
			}
			sample.Data["collection"] = sampleSubCol

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				f.Format(sample)
			}
		})
	}
}

func BenchmarkEntityFormatter_flatmap(b *testing.B) {
	numTargets := []int{0, 1, 2, 5, 10}
	depths := []int{1, 3, 7}
	for _, nTargets := range numTargets {
		for _, depth := range depths {
			extraFields := nTargets + depth*2
			extraSiblings := nTargets
			data, blacklist := benchmarkDeepStructure(nTargets, depth, extraFields, extraSiblings)
			sample := Response{
				Data:       data,
				IsComplete: true,
			}

			cmds := []interface{}{}
			for _, path := range blacklist {
				cmds = append(cmds, map[string]interface{}{
					"type": "del",
					"args": []interface{}{path},
				})
			}
			f := NewEntityFormatter(&config.Backend{
				ExtraConfig: config.ExtraConfig{
					Namespace: map[string]interface{}{
						flatmapKey: cmds,
					},
				},
			})
			b.Run(fmt.Sprintf("numTargets:%d,depth:%d,extraFields:%d,extraSiblings:%d", nTargets, depth, extraFields, extraSiblings), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					f.Format(sample)
				}
			})
		}
	}
}
