package graphql

import (
	"fmt"

	"github.com/devopsfaith/krakend/config"
)

func ExampleNewGraphQLParamExtractor() {
	cfg, err := GetOptions(config.ExtraConfig{
		Namespace: map[string]interface{}{
			"type":  OperationQuery,
			"query": "{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n",
			"variables": map[string]interface{}{
				"foo": "{foo}",
				"bar": "1234abc",
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	extractor := New(*cfg)

	body, err := extractor.BodyFromParams(map[string]string{
		"Foo": "foobar",
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))

	// output:
	// {"query":"{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n","variables":{"bar":"1234abc","foo":"foobar"}}
}
