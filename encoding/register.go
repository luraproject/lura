// SPDX-License-Identifier: Apache-2.0

package encoding

import (
	"io"

	"github.com/luraproject/lura/v2/register"
)

// GetRegister returns the package register
func GetRegister() *DecoderRegister {
	return decoders
}

type untypedRegister interface {
	Register(name string, v interface{})
	Get(name string) (interface{}, bool)
	Clone() map[string]interface{}
}

// DecoderRegister is the struct responsible of registering the decoder factories
type DecoderRegister struct {
	data untypedRegister
}

// Register adds a decoder factory to the register
func (r *DecoderRegister) Register(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error {
	r.data.Register(name, dec)
	return nil
}

// Get returns a decoder factory from the register by name. If no factory is found, it returns a JSON decoder factory
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
		JSON:      NewJSONDecoder,
		SAFE_JSON: NewSafeJSONDecoder,
		STRING:    NewStringDecoder,
		NOOP:      noOpDecoderFactory,
	}
)

func initDecoderRegister() *DecoderRegister {
	r := &DecoderRegister{data: register.NewUntyped()}
	for k, v := range defaultDecoders {
		r.Register(k, v)
	}
	return r
}
