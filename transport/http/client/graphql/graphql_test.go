package graphql

import (
	"fmt"
	"testing"

	"github.com/devopsfaith/krakend/config"
)

func TestNewGraphQLParamExtractor(t *testing.T) {
	cfg, err := GetOptions(config.ExtraConfig{
		Namespace: map[string]interface{}{
			"type":  OperationQuery,
			"query": "{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n}",
			"variables": map[string]interface{}{
				"foo": "{foo}",
				"bar": "1234abc",
			},
		},
	})
	if err != nil {
		t.Error(err)
		return
	}
	extractor := New(*cfg)

	body, err := extractor.ParamExtractor(map[string]string{
		"Foo": "foobar",
	})
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(string(body))
}

/*

/query/foo/{foo}


{
	"query": "{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n}",
	"variables": {
		"foo": "{foo}",
		"bar": "1234abc",
	}
}

*/
