package proxy

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/transport/http/client/graphql"
)

func TestNewGraphQLMiddleware_mutation(t *testing.T) {
	query := "mutation addAuthor($author: [AddAuthorInput!]!) {\n  addAuthor(input: $author) {\n    author {\n      id\n      name\n    }\n  }\n}\n"
	mw := NewGraphQLMiddleware(&config.Backend{
		ExtraConfig: config.ExtraConfig{
			graphql.Namespace: map[string]interface{}{
				"type":  "mutation",
				"query": query,
				"variables": map[string]interface{}{
					"author": map[string]interface{}{
						"name":  "A.N. Author",
						"dob":   "2000-01-01",
						"posts": []interface{}{},
					},
				},
			},
		},
	})

	expectedResponse := &Response{
		Data: map[string]interface{}{"foo": "bar"},
	}
	prxy := mw(func(ctx context.Context, req *Request) (*Response, error) {
		b, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
		var request graphql.GraphQLRequest
		if err := json.Unmarshal(b, &request); err != nil {
			return nil, err
		}
		return expectedResponse, nil
	})

	resp, err := prxy(context.Background(), &Request{
		Body: ioutil.NopCloser(strings.NewReader(`{
			"name": "foo",
			"dob": "bar"
		}`)),
		Params:  map[string]string{},
		Headers: map[string][]string{},
	})

	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(resp, expectedResponse) {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestNewGraphQLMiddleware_query(t *testing.T) {
	query := "{ q(func: uid(1)) { uid } }"
	mw := NewGraphQLMiddleware(&config.Backend{
		ExtraConfig: config.ExtraConfig{
			graphql.Namespace: map[string]interface{}{
				"type":  "query",
				"query": query,
				"variables": map[string]interface{}{
					"name":  "{foo}",
					"dob":   "{bar}",
					"posts": []interface{}{},
				},
			},
		},
	})

	expectedResponse := &Response{
		Data: map[string]interface{}{"foo": "bar"},
	}
	prxy := mw(func(ctx context.Context, req *Request) (*Response, error) {
		b, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
		var request graphql.GraphQLRequest
		if err := json.Unmarshal(b, &request); err != nil {
			return nil, err
		}
		return expectedResponse, nil
	})

	resp, err := prxy(context.Background(), &Request{
		Params: map[string]string{
			"Foo": "foo",
			"Bar": "bar",
		},
		Body:    ioutil.NopCloser(strings.NewReader("")),
		Headers: map[string][]string{},
	})

	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(resp, expectedResponse) {
		t.Errorf("unexpected response: %v", resp)
	}
}
