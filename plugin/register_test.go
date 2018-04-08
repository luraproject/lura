package plugin

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"plugin"
	"testing"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
	"github.com/devopsfaith/krakend/sd"
)

const samplePluginName = "samplePluginName"

func TestRegister_Register_ok(t *testing.T) {
	reg := NewRegister()
	p := dummyPlugin{
		content: plugin.Symbol(registrableDummy(1)),
	}
	if err := reg.Register(p); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestRegister_Register_ko(t *testing.T) {
	reg := NewRegister()
	p := dummyPlugin{
		err: errors.New("some error"),
	}
	if err := reg.Register(p); err == nil {
		t.Error("error expected")
		return
	}
}

func ExampleRegister_Register_ok() {
	reg := NewRegister()
	p := dummyPlugin{
		content: plugin.Symbol(registrableDummy(1)),
	}
	if err := reg.Register(p); err != nil {
		fmt.Println(err.Error())
	}
	// Output:
	// registrable 1 from plugin samplePluginName is registering its decoder components
	// registrable 1 from plugin samplePluginName is registering its SD components
	// registrable 1 from plugin samplePluginName is registering its components depending on external modules
}

func ExampleRegister_Register_unknownInterface() {
	reg := NewRegister()
	p := dummyPlugin{
		content: plugin.Symbol(1),
	}
	if err := reg.Register(p); err != nil {
		fmt.Println(err.Error())
	}
	// Output:
	// unknown registrable interface
}

type dummyPlugin struct {
	content plugin.Symbol
	err     error
}

func (d dummyPlugin) Lookup(name string) (plugin.Symbol, error) {
	if d.err != nil {
		return nil, d.err
	}

	if name != REGISTRABLE_VAR {
		return nil, fmt.Errorf("unknown symbol %s", name)
	}

	return d.content, nil
}

type registrableDummy int

func (r registrableDummy) RegisterDecoder(setter encoding.RegisterSetter) error {
	fmt.Println("registrable", r, "from plugin", samplePluginName, "is registering its decoder components")

	return setter.Register(samplePluginName, decoderFactory)
}

func (r registrableDummy) RegisterSD(setter sd.RegisterSetter) error {
	fmt.Println("registrable", r, "from plugin", samplePluginName, "is registering its SD components")

	return setter.Register(samplePluginName, subscriberFactory)
}

func (r registrableDummy) RegisterExternal(setter func(namespace, name string, v interface{})) error {
	fmt.Println("registrable", r, "from plugin", samplePluginName, "is registering its components depending on external modules")

	setter("namespace1", samplePluginName, func(x int) int { return 2 * x })
	return nil
}

func subscriberFactory(cfg *config.Backend) sd.Subscriber {
	fmt.Println("calling the SD factory:", samplePluginName)

	return sd.SubscriberFunc(func() ([]string, error) {
		fmt.Println("calling the subscriber:", samplePluginName)

		return cfg.Host, nil
	})
}

func decoderFactory(bool) encoding.Decoder {
	fmt.Println("calling the decoder factory:", samplePluginName)

	return func(reader io.Reader, _ *map[string]interface{}) error {
		fmt.Println("calling the decoder:", samplePluginName)

		d, err := ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
		fmt.Println("decoder:", samplePluginName, string(d))
		return nil
	}
}
