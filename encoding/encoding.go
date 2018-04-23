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

// Decoder is a function that reads from the reader and decodes it
// into an map of interfaces
type Decoder func(io.Reader, *map[string]interface{}) error

// DecoderFactory is a function that returns Decoder ready to deal with collections or with entities
// depending on the (isCollection) boolean value,
type DecoderFactory func(bool) func(io.Reader, *map[string]interface{}) error

// NOOP is the key for the NoOp encoding
const NOOP = "no-op"

// NoOpDecoder implements the Decoder interface
func NoOpDecoder(_ io.Reader, _ *map[string]interface{}) error { return nil }

func noOpDecoderFactory(_ bool) func(io.Reader, *map[string]interface{}) error { return NoOpDecoder }
