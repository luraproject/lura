package encoding

import (
	"io"

	"github.com/devopsfaith/krakend/register"
)

// Store is the struct responsible of registering the decoder factories
type Store struct {
	data UntypedRegisterer
}

// Set stores the received DecoderFactory into the register
func (r *Store) Set(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error {
	r.data.Set(name, dec)
	return nil
}

// Get returns a DecoderFactory depending on the received name. If no DecoderFactory is registered
// under the requested namespace, it will return the package defined fallback decoder facotry.
func (r *Store) Get(name string) func(bool) func(io.Reader, *map[string]interface{}) error {
	for _, n := range []string{name, fallbackDecoder} {
		if v, ok := r.data.Get(n); ok {
			if dec, ok := v.(func(bool) func(io.Reader, *map[string]interface{}) error); ok {
				return dec
			}
		}
	}

	return fallbackDecoderFactory
}

func initStore() *Store {
	r := register.NewUntyped()
	for k, v := range defaultDecoders {
		r.Set(k, v)
	}
	return &Store{r}
}
