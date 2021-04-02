package plugin

import (
	"context"
	"fmt"
	"github.com/devopsfaith/krakend/logging"
	"net/http"
	"plugin"
	"strings"

	krakendplugin "github.com/devopsfaith/krakend/plugin"
	"github.com/devopsfaith/krakend/register"
)

var serverRegister = register.New()

func RegisterHandler(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
) {
	serverRegister.Register(Namespace, name, handler)
}

func RegisterHandlerWithLogger(
	name string,
	handler func(context.Context, logging.Logger, map[string]interface{}, http.Handler) (http.Handler, error),
) {
	serverRegister.Register(Namespace, name, handler)
}

type Registerer interface {
	RegisterHandlers(func(
		name string,
		handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
	))
}

type RegistererWithLogger interface {
	RegisterHandlers(func(
		name string,
		handler func(context.Context, logging.Logger, map[string]interface{}, http.Handler) (http.Handler, error),
	))
}

type RegisterHandlerFunc func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)

type RegisterHandlerWithLoggerFunc func(
	name string,
	handler func(context.Context, logging.Logger, map[string]interface{}, http.Handler) (http.Handler, error),
)

func Load(path, pattern string, rcf RegisterHandlerFunc) (int, error) {
	plugins, err := krakendplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, rcf, RegisterHandlerWithLogger)
}

func load(plugins []string, rcf RegisterHandlerFunc, rclf RegisterHandlerWithLoggerFunc) (int, error) {
	errors := []error{}
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, rcf, rclf); err != nil {
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

func open(pluginName string, rcf RegisterHandlerFunc, rclf RegisterHandlerWithLoggerFunc) (err error) {
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
	r, err = p.Lookup("HandlerRegisterer")
	if err != nil {
		return
	}
	registerer, registererOk := r.(Registerer)
	registererWithLogger, registererWithLoggerOk := r.(RegistererWithLogger)
	if registererOk {
		registerer.RegisterHandlers(rcf)
	} else if registererWithLoggerOk {
		registererWithLogger.RegisterHandlers(rclf)
	} else {
		return fmt.Errorf("http-server-handler plugin loader: unknown type")
	}
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
