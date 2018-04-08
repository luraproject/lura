package main

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
	"github.com/devopsfaith/krakend/register"
	"github.com/devopsfaith/krakend/sd"
)

const pluginName = "supu"

var Registrable registrable

type registrable int

func (r *registrable) RegisterDecoder(setter encoding.RegisterSetter) error {
	fmt.Println("registrable", r, "from plugin", pluginName, "is registering its decoder components at", setter)

	return setter.Register(pluginName, decoderFactory)
}

func (r *registrable) RegisterSD(setter sd.RegisterSetter) error {
	fmt.Println("registrable", r, "from plugin", pluginName, "is registering its SD components at", setter)

	return setter.Register(pluginName, subscriberFactory)
}

func (r *registrable) RegisterExternal(setter *register.Namespaced) error {
	fmt.Println("registrable", r, "from plugin", pluginName, "is registering its components depending on external modules at", setter)

	setter.Register("namespace1", pluginName, doubleInt)
	return nil
}

func doubleInt(x int) int {
	return 2 * x
}

func subscriberFactory(cfg *config.Backend) sd.Subscriber {
	fmt.Println("calling the SD factory:", pluginName)

	return sd.SubscriberFunc(func() ([]string, error) {
		fmt.Println("calling the subscriber:", pluginName)

		return cfg.Host, nil
	})
}

func decoderFactory(bool) encoding.Decoder {
	fmt.Println("calling the decoder factory:", pluginName)

	return decoder
}

func decoder(reader io.Reader, _ *map[string]interface{}) error {
	fmt.Println("calling the decoder:", pluginName)

	d, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	fmt.Println("decoder:", pluginName, string(d))
	return nil
}

func main() {}
