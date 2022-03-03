// SPDX-License-Identifier: Apache-2.0

/*
Package encoding provides basic decoding implementations.

Decode decodes HTTP responses:

	resp, _ := http.Get("http://api.example.com/")
	...
	var data map[string]interface{}
	err := JSONDecoder(resp.Body, &data)

*/
package encoding

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

// Decoder is a function that reads from the reader and decodes it
// into an map of interfaces
type Decoder func(io.Reader, *map[string]interface{}) error

// DecoderFactory is a function that returns CollectionDecoder or an EntityDecoder
type DecoderFactory func(bool) func(io.Reader, *map[string]interface{}) error

// NOOP is the key for the NoOp encoding
const NOOP = "no-op"

// NoOpDecoder is a decoder that does nothing
func NoOpDecoder(_ io.Reader, _ *map[string]interface{}) error { return nil }

func noOpDecoderFactory(_ bool) func(io.Reader, *map[string]interface{}) error { return NoOpDecoder }

// JSON is the key for the json encoding
const JSON = "json"

// NewJSONDecoder returns the right JSON decoder
func NewJSONDecoder(isCollection bool) func(io.Reader, *map[string]interface{}) error {
	if isCollection {
		return JSONCollectionDecoder
	}
	return JSONDecoder
}

// JSONDecoder decodes a json message into a map
func JSONDecoder(r io.Reader, v *map[string]interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(v)
}

// JSONCollectionDecoder decodes a json collection and returns a map with the array at the 'collection' key
func JSONCollectionDecoder(r io.Reader, v *map[string]interface{}) error {
	var collection []interface{}
	d := json.NewDecoder(r)
	d.UseNumber()
	if err := d.Decode(&collection); err != nil {
		return err
	}
	*(v) = map[string]interface{}{"collection": collection}
	return nil
}

// SAFE_JSON is the key for the json encoding
const SAFE_JSON = "safejson"

// NewSafeJSONDecoder returns the universal json decoder
func NewSafeJSONDecoder(_ bool) func(io.Reader, *map[string]interface{}) error {
	return SafeJSONDecoder
}

// SafeJSONDecoder decodes both json objects and collections
func SafeJSONDecoder(r io.Reader, v *map[string]interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	var t interface{}
	if err := d.Decode(&t); err != nil {
		return err
	}
	switch tt := t.(type) {
	case map[string]interface{}:
		*v = tt
	case []interface{}:
		*v = map[string]interface{}{"collection": tt}
	default:
		*v = map[string]interface{}{"result": tt}
	}
	return nil
}

// STRING is the key for the string encoding
const STRING = "string"

// NewStringDecoder returns a String decoder
func NewStringDecoder(_ bool) func(io.Reader, *map[string]interface{}) error {
	return StringDecoder
}

// StringDecoder returns a map with the content of the reader under the key 'content'
func StringDecoder(r io.Reader, v *map[string]interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	*(v) = map[string]interface{}{"content": string(data)}
	return nil
}
