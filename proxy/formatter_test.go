package proxy

import "testing"

func TestEntityFormatter_newWhitelistingFilter(t *testing.T) {
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
				"d": "tupu",
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter("", []string{"supu", "a.b", "a.c", "foo.unknown"}, []string{}, "", map[string]string{})
	result := f.Format(sample)
	if v, ok := result.Data["supu"]; !ok || v != expected.Data["supu"] {
		t.Errorf("The formatter returned an unexpected result for the field supu: %v\n", result)
	}
	v, ok := result.Data["a"]
	if !ok {
		t.Errorf("The formatter returned an unexpected result for the fields a.b & a.c: %v\n", result)
	}
	tmp := v.(map[string]interface{})
	if b, okk := tmp["b"]; !okk || !b.(bool) {
		t.Errorf("The formatter returned an unexpected result for the field a.b: %v\n", result)
	}
	if c, okk := tmp["c"]; !okk || c.(int) != 42 {
		t.Errorf("The formatter returned an unexpected result for the field a.c: %v\n", result)
	}
	if len(tmp) != 2 {
		t.Errorf("The formatter returned an unexpected result size for the field a: %v\n", result)
	}
	if len(result.Data) != 2 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_newblacklistingFilter(t *testing.T) {
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
				"d": "tupu",
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"tupu": false,
			"foo":  "bar",
			"a": map[string]interface{}{
				"d": "tupu",
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter("", []string{}, []string{"supu", "a.b", "a.c", "foo.unknown"}, "", map[string]string{})
	result := f.Format(sample)
	if v, ok := result.Data["tupu"]; !ok || v != expected.Data["tupu"] {
		t.Errorf("The formatter returned an unexpected result for the field tupu: %v\n", result)
	}
	if v, ok := result.Data["foo"]; !ok || v != expected.Data["foo"] {
		t.Errorf("The formatter returned an unexpected result for the field foo: %v\n", result)
	}
	v, ok := result.Data["a"]
	if !ok {
		t.Errorf("The formatter returned an unexpected result for the field a.d: %v\n", result)
	}
	tmp := v.(map[string]interface{})
	if d, okk := tmp["d"]; !okk || d != "tupu" {
		t.Errorf("The formatter returned an unexpected result for the field a.d: %v\n", result)
	}
	if len(tmp) != 1 {
		t.Errorf("The formatter returned an unexpected result size for the field a: %v\n", result)
	}
	if len(result.Data) != 3 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_grouping(t *testing.T) {
	preffix := "group1"
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			preffix: map[string]interface{}{
				"supu": 42,
				"tupu": false,
				"foo":  "bar",
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter("", []string{}, []string{}, preffix, map[string]string{})
	result := f.Format(sample)
	if len(result.Data) != 1 || result.IsComplete != expected.IsComplete {
		t.Fail()
	}
	if _, ok := result.Data[preffix]; !ok {
		t.Fail()
	}
	group := result.Data[preffix].(map[string]interface{})
	for k, expectedValue := range expected.Data[preffix].(map[string]interface{}) {
		if v, ok := group[k]; !ok || v != expectedValue {
			t.Fail()
		}
	}
}

func TestEntityFormatter_mapping(t *testing.T) {
	mapping := map[string]string{"supu": "SUPUUUUU", "tupu": "TUPUUUUU", "a.b": "a.BOOOOO"}

	sub := map[string]interface{}{
		"b": true,
		"c": 42,
		"d": "tupu",
	}
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
			"a":    sub,
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"SUPUUUUU": 42,
			"TUPUUUUU": false,
			"foo":      "bar",
			"a":        sub,
		},
		IsComplete: true,
	}
	f := NewEntityFormatter("", []string{}, []string{}, "", mapping)
	result := f.Format(sample)

	if len(result.Data) != 4 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result.Data)
	}
	for k, expectedValue := range expected.Data {
		if k == "a" {
			continue
		}
		if v, ok := result.Data[k]; !ok || v != expectedValue {
			t.Errorf("The formatter returned an unexpected result for the key %s: %v\n", k, v)
		}
	}
	group := result.Data["a"].(map[string]interface{})
	for k, expectedValue := range expected.Data["a"].(map[string]interface{}) {
		if v, ok := group[k]; !ok || v != expectedValue {
			t.Errorf("The formatter returned an unexpected result for the key %s: %v\n", k, v)
		}
	}

	if len(group) != 3 {
		t.Errorf("The formatter returned an unexpected result size for the subentity: %v\n", group)
	}
}

func TestEntityFormatter_targeting(t *testing.T) {
	target := "group1"
	sub := map[string]interface{}{
		"b": true,
		"c": 42,
		"d": "tupu",
	}
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
			target: sub,
		},
		IsComplete: true,
	}
	expected := Response{
		Data:       sub,
		IsComplete: true,
	}
	f := NewEntityFormatter(target, []string{}, []string{}, "", map[string]string{})
	result := f.Format(sample)
	if len(result.Data) != 3 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
	for k, expectedValue := range expected.Data {
		if v, ok := result.Data[k]; !ok || v != expectedValue {
			t.Errorf("The formatter returned an unexpected result for the key %s: %v\n", k, v)
		}
	}
}

func TestEntityFormatter_targetingUnknownFields(t *testing.T) {
	target := "group1"
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(target, []string{}, []string{}, "", map[string]string{})
	result := f.Format(sample)
	if len(result.Data) != 0 || result.IsComplete != sample.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_targetingNonObjects(t *testing.T) {
	target := "group1"
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
			target: false,
		},
		IsComplete: true,
	}
	f := NewEntityFormatter(target, []string{}, []string{}, "", map[string]string{})
	result := f.Format(sample)
	if len(result.Data) != 0 || result.IsComplete != sample.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}

func TestEntityFormatter_altogether(t *testing.T) {
	sample := Response{
		Data: map[string]interface{}{
			"supu": 42,
			"tupu": false,
			"foo":  "bar",
			"a": map[string]interface{}{
				"b": true,
				"c": 42,
				"d": "tupu",
			},
		},
		IsComplete: true,
	}
	expected := Response{
		Data: map[string]interface{}{
			"group": map[string]interface{}{
				"D": "tupu",
			},
		},
		IsComplete: true,
	}
	f := NewEntityFormatter("a", []string{"d"}, []string{}, "group", map[string]string{"d": "D"})
	result := f.Format(sample)
	v, ok := result.Data["group"]
	if !ok {
		t.Errorf("The formatter returned an unexpected result for the field group.D: %v\n", result)
	}
	tmp := v.(map[string]interface{})
	if d, okk := tmp["D"]; !okk || d != "tupu" {
		t.Errorf("The formatter returned an unexpected result for the field group.D: %v\n", result)
	}
	if len(tmp) != 1 {
		t.Errorf("The formatter returned an unexpected result size for the field group: %v\n", result)
	}
	if len(result.Data) != 1 || result.IsComplete != expected.IsComplete {
		t.Errorf("The formatter returned an unexpected result size: %v\n", result)
	}
}
