package proxy

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/devopsfaith/krakend/config"
)

func BenchmarkEntityFormatter_whitelistingFilter(b *testing.B) {
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
				f := NewEntityFormatter(&config.Backend{Whitelist: testCase})
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					f.Format(sample)
				}
			})
		}
	}

}

func benchmarkDeepChilds(depth int, extraSiblings int) map[string]interface{} {
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

func benchmarkDeepStructure(numTargets int, targetDepth int, extraFields int, extraSiblings int) (map[string]interface{}, []string) {
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

func BenchmarkEntityFormatter_deepWhitelistingFilter(b *testing.B) {
	numTargets := []int{0, 1, 2, 5, 10}
	depths := []int{1, 3, 7}
	for _, nTargets := range numTargets {
		for _, depth := range depths {
			extraFields := nTargets + depth*2
			extraSiblings := nTargets
			data, whitelist := benchmarkDeepStructure(nTargets, depth, extraFields, extraSiblings)
			sample := Response{
				Data:       data,
				IsComplete: true,
			}
			f := NewEntityFormatter(&config.Backend{Whitelist: whitelist})
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

func BenchmarkEntityFormatter_blacklistingFilter(b *testing.B) {
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
				f := NewEntityFormatter(&config.Backend{Blacklist: testCase})
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
