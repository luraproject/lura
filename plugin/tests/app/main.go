package main

import (
	"bytes"
	"fmt"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/plugin"
)

const pluginName = "supu"

func main() {
	register := plugin.NewRegister()

	fmt.Println(plugin.Load(config.Plugin{
		Folder:  "../plugins/",
		Pattern: ".so",
	}, register))

	time.Sleep(5 * time.Second)

	checkDecoder(register)
	checkExternal(register)
}

func checkDecoder(register *plugin.Register) {
	decoderFactory := register.Decoder.Get(pluginName)
	decoder := decoderFactory(false)
	d := bytes.NewBufferString("something")
	m := &map[string]interface{}{}
	if err := decoder(d, m); err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	fmt.Println("encoding ok!")
}

func checkExternal(register *plugin.Register) {
	n1Register, ok := register.External.Get("namespace1")
	if !ok {
		fmt.Println("namespace1 not registered")
		return
	}
	v, ok := n1Register.Get(pluginName)
	if !ok {
		fmt.Println(pluginName, "not registered into namespace1")
		return
	}
	f, ok := v.(func(int) int)
	if !ok {
		fmt.Println("unexpected registerd component into namespace1", v)
		return
	}
	fmt.Println("f(2) =", f(2))
}
