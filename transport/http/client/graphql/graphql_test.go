// SPDX-License-Identifier: Apache-2.0

package graphql

import (
	"fmt"
	"os"
	"strings"

	"github.com/luraproject/lura/v2/config"
)

func ExampleExtractor() {
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

	{
		fmt.Println("BodyFromParams")
		body, err := extractor.BodyFromParams(map[string]string{
			"Foo": "foobar",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(body))
	}

	{
		fmt.Println("QueryFromParams")
		query, err := extractor.QueryFromParams(map[string]string{
			"Foo": "foobar",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(query.Encode())
	}

	{
		fmt.Println("BodyFromBody")
		body, err := extractor.BodyFromBody(strings.NewReader(`{
		"foo": "foobar",
		"foo1": "foobar"
	}`))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(body))
	}

	{
		fmt.Println("QueryFromBody")
		query, err := extractor.QueryFromBody(strings.NewReader(`{
		"foo": "foobar",
		"foo1": "foobar"
	}`))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(query.Encode())
	}

	// output:
	// BodyFromParams
	// {"query":"{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n","variables":{"bar":"1234abc","foo":"foobar"}}
	// QueryFromParams
	// query=%7B%0A++find_follower%28func%3A+uid%28%220x3%22%29%29+%7B%0A++++name+%0A++++%7D%0A++%7D%0A&variables=%7B%22bar%22%3A%221234abc%22%2C%22foo%22%3A%22foobar%22%7D
	// BodyFromBody
	// {"query":"{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n","variables":{"bar":"1234abc","foo":"foobar","foo1":"foobar"}}
	// QueryFromBody
	// query=%7B%0A++find_follower%28func%3A+uid%28%220x3%22%29%29+%7B%0A++++name+%0A++++%7D%0A++%7D%0A&variables=%7B%22bar%22%3A%221234abc%22%2C%22foo%22%3A%22foobar%22%2C%22foo1%22%3A%22foobar%22%7D

}

func ExampleExtractor_fromFile() {
	os.WriteFile(".graphql_query.txt", []byte("{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n"), 0664)
	defer os.Remove(".graphql_query.txt")

	cfg, err := GetOptions(config.ExtraConfig{
		Namespace: map[string]interface{}{
			"type":       OperationQuery,
			"query_path": ".graphql_query.txt",
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

	{
		fmt.Println("BodyFromParams")
		body, err := extractor.BodyFromParams(map[string]string{
			"Foo": "foobar",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(body))
	}

	{
		fmt.Println("QueryFromParams")
		query, err := extractor.QueryFromParams(map[string]string{
			"Foo": "foobar",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(query.Encode())
	}

	{
		fmt.Println("BodyFromBody")
		body, err := extractor.BodyFromBody(strings.NewReader(`{
		"foo": "foobar",
		"foo1": "foobar"
	}`))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(body))
	}

	{
		fmt.Println("QueryFromBody")
		query, err := extractor.QueryFromBody(strings.NewReader(`{
		"foo": "foobar",
		"foo1": "foobar"
	}`))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(query.Encode())
	}

	// output:
	// BodyFromParams
	// {"query":"{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n","variables":{"bar":"1234abc","foo":"foobar"}}
	// QueryFromParams
	// query=%7B%0A++find_follower%28func%3A+uid%28%220x3%22%29%29+%7B%0A++++name+%0A++++%7D%0A++%7D%0A&variables=%7B%22bar%22%3A%221234abc%22%2C%22foo%22%3A%22foobar%22%7D
	// BodyFromBody
	// {"query":"{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n","variables":{"bar":"1234abc","foo":"foobar","foo1":"foobar"}}
	// QueryFromBody
	// query=%7B%0A++find_follower%28func%3A+uid%28%220x3%22%29%29+%7B%0A++++name+%0A++++%7D%0A++%7D%0A&variables=%7B%22bar%22%3A%221234abc%22%2C%22foo%22%3A%22foobar%22%2C%22foo1%22%3A%22foobar%22%7D

}

func ExampleExtractor_noReplacement() {
	cfg, err := GetOptions(config.ExtraConfig{
		Namespace: map[string]interface{}{
			"type":  OperationQuery,
			"query": "{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n",
			"variables": map[string]interface{}{
				"bar": "1234abc",
			},
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	extractor := New(*cfg)

	{
		fmt.Println("BodyFromParams")
		body, err := extractor.BodyFromParams(map[string]string{
			"Foo": "foobar",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(body))
	}

	{
		fmt.Println("QueryFromParams")
		query, err := extractor.QueryFromParams(map[string]string{
			"Foo": "foobar",
		})
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(query.Encode())
	}

	{
		fmt.Println("BodyFromBody")
		body, err := extractor.BodyFromBody(strings.NewReader(`{
		"foo": "foobar",
		"foo1": "foobar"
	}`))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(string(body))
	}

	{
		fmt.Println("QueryFromBody")
		query, err := extractor.QueryFromBody(strings.NewReader(`{
		"foo": "foobar",
		"foo1": "foobar"
	}`))
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(query.Encode())
	}

	// output:
	// BodyFromParams
	// {"query":"{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n","variables":{"bar":"1234abc"}}
	// QueryFromParams
	// query=%7B%0A++find_follower%28func%3A+uid%28%220x3%22%29%29+%7B%0A++++name+%0A++++%7D%0A++%7D%0A&variables=%7B%22bar%22%3A%221234abc%22%7D
	// BodyFromBody
	// {"query":"{\n  find_follower(func: uid(\"0x3\")) {\n    name \n    }\n  }\n","variables":{"bar":"1234abc","foo":"foobar","foo1":"foobar"}}
	// QueryFromBody
	// query=%7B%0A++find_follower%28func%3A+uid%28%220x3%22%29%29+%7B%0A++++name+%0A++++%7D%0A++%7D%0A&variables=%7B%22bar%22%3A%221234abc%22%2C%22foo%22%3A%22foobar%22%2C%22foo1%22%3A%22foobar%22%7D

}
