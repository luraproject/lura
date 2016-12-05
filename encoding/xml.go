package encoding

import (
	"encoding/xml"
	"io"
)

// XMLDecoder implements the Decoder interface
func XMLDecoder(r io.Reader, v *map[string]interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}
