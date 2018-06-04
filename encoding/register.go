package encoding

import (
	"io"

	"github.com/devopsfaith/krakend/register"
)

// GetRegister returns the package register
func GetRegister() *DecoderRegister {
	return decoders
}

// DecoderRegister is the struct responsible of registering the decoder factories
type DecoderRegister struct {
	data register.Untyped
}

// Register implements the RegisterSetter interface
func (r *DecoderRegister) Register(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error {
	r.data.Register(name, dec)
	return nil
}

// Get implements the RegisterGetter interface
func (r *DecoderRegister) Get(name string) func(bool) func(io.Reader, *map[string]interface{}) error {
	for _, n := range []string{name, JSON} {
		if v, ok := r.data.Get(n); ok {
			if dec, ok := v.(func(bool) func(io.Reader, *map[string]interface{}) error); ok {
				return dec
			}
		}
	}
	return NewJSONDecoder
}

var (
	decoders        = initDecoderRegister()
	defaultDecoders = map[string]func(bool) func(io.Reader, *map[string]interface{}) error{
		JSON:   NewJSONDecoder,
		STRING: NewStringDecoder,
		NOOP:   noOpDecoderFactory,
	}
)

func initDecoderRegister() *DecoderRegister {
	r := &DecoderRegister{register.NewUntyped()}
	for k, v := range defaultDecoders {
		r.Register(k, v)
	}
	return r
}
