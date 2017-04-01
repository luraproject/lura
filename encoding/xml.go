package encoding

import (
	"encoding/xml"
	"io"
)

// XMLDecoder implements the Decoder interface
func XMLDecoder(r io.Reader, v *map[string]interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}

// XMLCollectionDecoder implements the Decoder interface over a collection
func XMLCollectionDecoder(r io.Reader, v *map[string]interface{}) error {
	var collection []interface{}
	if err := xml.NewDecoder(r).Decode(&collection); err != nil {
		return err
	}
	*(v) = map[string]interface{}{"collection": collection}
	return nil
}
