package proxy

import "testing"

func TestRequestGeneratePath(t *testing.T) {
	r := Request{
		Method: "GET",
		Params: map[string]string{
			"Supu": "42",
			"Tupu": "false",
			"Foo":  "bar",
		},
	}

	for i, testCase := range [][]string{
		{"/a/{{.Supu}}", "/a/42"},
		{"/a?b={{.Tupu}}", "/a?b=false"},
		{"/a/{{.Supu}}/foo/{{.Foo}}", "/a/42/foo/bar"},
		{"/a", "/a"},
	} {
		r.GeneratePath(testCase[0])
		if r.Path != testCase[1] {
			t.Errorf("%d: want %s, have %s", i, testCase[1], r.Path)
		}
	}
}
