/*
Package encoding provides Decoding implementations.

Decode decodes HTTP responses:

	resp, _ := http.Get("http://api.example.com/")
	...
	var data map[string]interface{}
	err := JSONDecoder(resp.Body, &data)

*/
package encoding

import "io"

// A Decoder is a function that reads from the reader and decodes it
// into an map of interfaces
type Decoder func(io.Reader, *map[string]interface{}) error

// A DecoderFactory is a function that returns CollectionDecoder or an EntityDecoder
type DecoderFactory func(bool) func(io.Reader, *map[string]interface{}) error

// Deprecated: Register is deprecated
func Register(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error {
	return decoders.Register(name, dec)
}

// Deprecated: Get is deprecated
func Get(name string) DecoderFactory {
	return decoders.Get(name)
}

// NOOP is the key for the NoOp encoding
const NOOP = "no-op"

// NoOpDecoder implements the Decoder interface
func NoOpDecoder(_ io.Reader, _ *map[string]interface{}) error { return nil }

func noOpDecoderFactory(_ bool) func(io.Reader, *map[string]interface{}) error { return NoOpDecoder }
