package main

import (
	"bytes"
	"fmt"
	"reflect"
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
	checkSD(register)
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

func checkSD(register *plugin.Register) {
	sdFactory := register.SD.Get(pluginName)
	cfg := &config.Backend{Host: []string{"a", "b"}}
	subscriber := sdFactory(cfg)
	hosts, err := subscriber.Hosts()
	if err != nil {
		fmt.Println("error:", err.Error())
		return
	}
	if !reflect.DeepEqual(hosts, cfg.Host) {
		fmt.Println("unexpected set of hosts:", hosts)
		return
	}
	fmt.Println("sd ok!")
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
