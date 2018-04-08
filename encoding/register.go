package encoding

import "sync"

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

var decoders = &DecoderRegister{
	data: map[string]DecoderFactory{
		JSON:   NewJSONDecoder,
		STRING: NewStringDecoder,
	},
	mutex: &sync.RWMutex{},
}

type DecoderRegister struct {
	data  map[string]DecoderFactory
	mutex *sync.RWMutex
}

// Register implements the RegisterSetter interface
func (r *DecoderRegister) Register(name string, dec DecoderFactory) error {
	r.mutex.Lock()
	r.data[name] = dec
	r.mutex.Unlock()
	return nil
}

// Get implements the RegisterGetter interface
func (r *DecoderRegister) Get(name string) DecoderFactory {
	for _, n := range []string{name, JSON} {
		if dec, ok := r.data[n]; ok {
			return dec
		}
	}
	return NewJSONDecoder
}
