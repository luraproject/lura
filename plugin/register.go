package plugin

import (
	"fmt"
	"io"

	"github.com/devopsfaith/krakend/encoding"
	"github.com/devopsfaith/krakend/register"
	"github.com/devopsfaith/krakend/sd"
)

// REGISTRABLE_VAR is the name to lookup after loading the plugin for the module registering
const REGISTRABLE_VAR = "Registrable"

// Register contains all the registers required by the framework and the external modules
type Register struct {
	Decoder  *encoding.DecoderRegister
	SD       *sd.Register
	External *register.Namespaced
}

// NewRegister returns a new register to be used by the plugin loader
func NewRegister() *Register {
	return &Register{
		Decoder:  encoding.GetRegister(),
		SD:       sd.GetRegister(),
		External: register.New(),
	}
}

// Register registers the received plugin in the propper internal registers
func (r *Register) Register(p Plugin) error {
	x, err := p.Lookup(REGISTRABLE_VAR)
	if err != nil {
		fmt.Println("unable to find the registrable symbol:", err.Error())
		return err
	}

	totalRegistrations := 0

	if registrable, ok := x.(RegistrableDecoder); ok {
		err = registrable.RegisterDecoder(r.Decoder.Register)
		totalRegistrations++
	}

	if registrable, ok := x.(RegistrableExternal); ok {
		err = registrable.RegisterExternal(r.External.Register)
		totalRegistrations++
	}

	if totalRegistrations == 0 {
		fmt.Println("unknown registrable interface")
	}

	return nil
}

// RegistrableDecoder defines the interface the encoding plugins should implement
// in order to be able to register themselves
type RegistrableDecoder interface {
	RegisterDecoder(func(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error) error
}

// RegistrableExternal defines the interface the external plugins should implement
// in order to be able to register themselves
type RegistrableExternal interface {
	RegisterExternal(func(namespace, name string, v interface{})) error
}
