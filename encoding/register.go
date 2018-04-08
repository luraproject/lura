package encoding

import "github.com/devopsfaith/krakend/register"

// RegisterSetter registers the decoder factory with the given name
type RegisterSetter interface {
	Register(name string, dec DecoderFactory) error
}

// RegisterGetter returns the decoder factory by name. If there is no factory with the received name
// it returns the JSON decoder factory
type RegisterGetter interface {
	Get(name string) DecoderFactory
}

// GetRegister returns the package register
func GetRegister() *DecoderRegister {
	return decoders
}

// DecoderRegister is the struct responsible of registering the decoder factories
type DecoderRegister struct {
	data register.Untyped
}

// Register implements the RegisterSetter interface
func (r *DecoderRegister) Register(name string, dec DecoderFactory) error {
	r.data.Register(name, dec)
	return nil
}

// Get implements the RegisterGetter interface
func (r *DecoderRegister) Get(name string) DecoderFactory {
	for _, n := range []string{name, JSON} {
		if v, ok := r.data.Get(n); ok {
			if dec, ok := v.(DecoderFactory); ok {
				return dec
			}
		}
	}
	return NewJSONDecoder
}

var (
	decoders        = initDecoderRegister()
	defaultDecoders = map[string]DecoderFactory{
		JSON:   NewJSONDecoder,
		STRING: NewStringDecoder,
	}
)

func initDecoderRegister() *DecoderRegister {
	r := &DecoderRegister{register.NewUntyped()}
	for k, v := range defaultDecoders {
		r.Register(k, v)
	}
	return r
}
