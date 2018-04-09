package encoding

import (
	"io"
	"io/ioutil"
)

// STRING is the key for the string encoding
const STRING = "string"

// NewStringDecoder return a String decoder
func NewStringDecoder(_ bool) func(io.Reader, *map[string]interface{}) error {
	return StringDecoder
}

// StringDecoder implements the Decoder interface
func StringDecoder(r io.Reader, v *map[string]interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	*(v) = map[string]interface{}{"content": string(data)}
	return nil
}
