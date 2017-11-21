package encoding

import (
	"io"
	"io/ioutil"
)

// RAW is the key for the nop encoding
const RAW = "raw"

// NewRawDecoder return a Nop/Raw decoder
func NewRawDecoder(_ bool) Decoder {
	return RawDecoder
}

// RawDecoder implements the Decoder interface
func RawDecoder(r io.Reader, v *map[string]interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	*(v) = map[string]interface{}{"content": string(data)}
	return nil
}
