package main

import (
	"fmt"
	"io"
	"io/ioutil"
)

const pluginName = "supu"

var Registrable registrable

type registrable int

func (r *registrable) RegisterDecoder(setter func(name string, dec func(bool) func(io.Reader, *map[string]interface{}) error) error) error {
	fmt.Println("registrable", r, "from plugin", pluginName, "is registering its decoder components")

	return setter(pluginName, decoderFactory)
}

func (r *registrable) RegisterExternal(setter func(namespace, name string, v interface{})) error {
	fmt.Println("registrable", r, "from plugin", pluginName, "is registering its components depending on external modules")

	setter("namespace1", pluginName, doubleInt)
	return nil
}

func doubleInt(x int) int {
	return 2 * x
}

func decoderFactory(bool) func(reader io.Reader, _ *map[string]interface{}) error {
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
