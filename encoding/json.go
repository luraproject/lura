package encoding

import (
	"encoding/json"
	"io"
)

// JSON is the key for the json encoding
const JSON = "json"

// NewJSONDecoder return the right JSON decoder
func NewJSONDecoder(isCollection bool) func(io.Reader, *map[string]interface{}) error {
	if isCollection {
		return JSONCollectionDecoder
	}
	return JSONDecoder
}

// JSONDecoder implements the Decoder interface
func JSONDecoder(r io.Reader, v *map[string]interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(v)
}

// JSONCollectionDecoder implements the Decoder interface over a collection
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

// NewJSONDecoder return the right JSON decoder
func NewSafeJSONDecoder(isCollection bool) func(io.Reader, *map[string]interface{}) error {
	return SafeJSONDecoder
}

// JSONDecoder implements the Decoder interface
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
