package plugin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"plugin"
	"strings"

	"github.com/devopsfaith/krakend/register"
)

var clientRegister = register.New()

func RegisterClient(
	name string,
	handler func(ctx context.Context, extra map[string]interface{}) (http.Handler, error),
) {
	clientRegister.Register(Namespace, name, handler)
}

type Registerer interface {
	RegisterClients(func(
		name string,
		handler func(ctx context.Context, extra map[string]interface{}) (http.Handler, error),
	))
}

type RegisterClientFunc func(
	name string,
	handler func(ctx context.Context, extra map[string]interface{}) (http.Handler, error),
)

func Load(path, pattern string, rcf RegisterClientFunc) (int, error) {
	plugins, err := scan(path, pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, rcf)
}

func scan(folder, pattern string) ([]string, error) {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return []string{}, err
	}

	plugins := []string{}
	for _, file := range files {
		if !file.IsDir() && strings.Contains(file.Name(), pattern) {
			plugins = append(plugins, folder+file.Name())
		}
	}

	return plugins, nil
}

func load(plugins []string, rcf RegisterClientFunc) (int, error) {
	errors := []error{}
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, rcf); err != nil {
			errors = append(errors, fmt.Errorf("opening plugin %d (%s): %s", k, pluginName, err.Error()))
			continue
		}
		loadedPlugins++
	}

	if len(errors) > 0 {
		return loadedPlugins, loaderError{errors}
	}
	return loadedPlugins, nil
}

func open(pluginName string, rcf RegisterClientFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	var p Plugin
	p, err = pluginOpener(pluginName)
	if err != nil {
		return
	}
	var r interface{}
	r, err = p.Lookup("GRPCRegisterer")
	if err != nil {
		return
	}
	registerer, ok := r.(Registerer)
	if !ok {
		return fmt.Errorf("gRPC plugin loader: unknown type")
	}
	registerer.RegisterClients(rcf)
	return
}

// Plugin is the interface of the loaded plugins
type Plugin interface {
	Lookup(name string) (plugin.Symbol, error)
}

// pluginOpener keeps the plugin open function in a var for easy testing
var pluginOpener = defaultPluginOpener

func defaultPluginOpener(name string) (Plugin, error) {
	return plugin.Open(name)
}

type loaderError struct {
	errors []error
}

// Error implements the error interface
func (l loaderError) Error() string {
	msgs := make([]string, len(l.errors))
	for i, err := range l.errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("plugin loader found %d error(s): \n%s", len(msgs), strings.Join(msgs, "\n"))
}
