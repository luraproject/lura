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
type DecoderFactory func(bool) Decoder

var decoders = map[string]DecoderFactory{
	JSON:   NewJSONDecoder,
	STRING: NewStringDecoder,
}

// Register registers the decoder factory with the given name
func Register(name string, dec DecoderFactory) error {
	decoders[name] = dec
	return nil
}

// Get returns (from the register) the decoder factory by name. If there is no factory with the received name
// it returns the JSON decoder factory
func Get(name string) DecoderFactory {
	for _, n := range []string{name, JSON} {
		if dec, ok := decoders[n]; ok {
			return dec
		}
	}
	return NewJSONDecoder
}
