package encoding

import "io"

// Deprecated: Register is deprecated
func Register(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error {
	return GetRegister().Set(name, dec)
}

// Deprecated: Get is deprecated
func Get(name string) DecoderFactory {
	return GetRegister().Get(name)
}

// Registerer defines the interface of the package registerer
type Registerer interface {
	Set(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error
	Get(name string) func(bool) func(io.Reader, *map[string]interface{}) error
}

// GetRegister returns the package register
func GetRegister() Registerer {
	return decoders
}

// UntypedRegisterer defines the interface of the internal registerer
type UntypedRegisterer interface {
	Set(name string, v interface{})
	Get(name string) (interface{}, bool)
	Clone() map[string]interface{}
}

var (
	defaultDecoders = map[string]func(bool) func(io.Reader, *map[string]interface{}) error{
		JSON:   NewJSONDecoder,
		STRING: NewStringDecoder,
		NOOP:   noOpDecoderFactory,
	}
	decoders               = initStore()
	fallbackDecoder        = JSON
	fallbackDecoderFactory = NewJSONDecoder
)
