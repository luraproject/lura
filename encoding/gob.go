package encoding

import (
	"encoding/gob"
	"io"
)

// GOB is the key for the gob encoding
const GOB = "gob"

// NewGobDecoder return the right Gob decoder
func NewGobDecoder(isCollection bool) func(io.Reader, *map[string]interface{}) error {
	if isCollection {
		return GobCollectionDecoder
	}
	return GobDecoder
}

// GobDecoder implements the Decoder interface
func GobDecoder(r io.Reader, v *map[string]interface{}) error {
	return gob.NewDecoder(r).Decode(v)
}

// GobCollectionDecoder implements the Decoder interface over a collection
func GobCollectionDecoder(r io.Reader, v *map[string]interface{}) error {
	var collection []interface{}
	if err := gob.NewDecoder(r).Decode(&collection); err != nil {
		return err
	}
	*(v) = map[string]interface{}{"collection": collection}
	return nil
}
