package proxy

import (
	"fmt"
	"testing"
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
				f := NewEntityFormatter("", testCase, []string{}, "", map[string]string{})
				b.ResetTimer()
				b.ReportAllocs()
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
				f := NewEntityFormatter("", []string{}, testCase, "", map[string]string{})
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
			f := NewEntityFormatter("", []string{}, []string{}, preffix, map[string]string{})
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
				f := NewEntityFormatter("", []string{}, []string{}, "", testCase)
				b.ResetTimer()
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					f.Format(sample)
				}
			})
		}
	}
}
