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
type Decoder func(r io.Reader, v *map[string]interface{}) error
