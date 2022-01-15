// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"fmt"
	"net/http"
	"plugin"
	"strings"

	"github.com/luraproject/lura/v2/logging"
	luraplugin "github.com/luraproject/lura/v2/plugin"
	"github.com/luraproject/lura/v2/register"
)

var serverRegister = register.New()

func RegisterHandler(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
) {
	serverRegister.Register(Namespace, name, handler)
}

type Registerer interface {
	RegisterHandlers(func(
		name string,
		handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
	))
}

type LoggerRegisterer interface {
	RegisterLogger(interface{})
}

type RegisterHandlerFunc func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)

func Load(path, pattern string, rcf RegisterHandlerFunc) (int, error) {
	return LoadWithLogger(path, pattern, rcf, nil)
}

func LoadWithLogger(path, pattern string, rcf RegisterHandlerFunc, logger logging.Logger) (int, error) {
	plugins, err := luraplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, rcf, logger)
}

func load(plugins []string, rcf RegisterHandlerFunc, logger logging.Logger) (int, error) {
	errors := []error{}
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, rcf, logger); err != nil {
			errors = append(errors, fmt.Errorf("plugin #%d (%s): %s", k, pluginName, err.Error()))
			continue
		}
		loadedPlugins++
	}

	if len(errors) > 0 {
		return loadedPlugins, loaderError{errors: errors}
	}
	return loadedPlugins, nil
}

func open(pluginName string, rcf RegisterHandlerFunc, logger logging.Logger) (err error) {
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
	registerer, ok := r.(Registerer)
	if !ok {
		return fmt.Errorf("http-server-handler plugin loader: unknown type")
	}

	if logger != nil {
		if lr, ok := r.(LoggerRegisterer); ok {
			lr.RegisterLogger(logger)
		}
	}

	registerer.RegisterHandlers(rcf)
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

func (l loaderError) Len() int {
	return len(l.errors)
}

func (l loaderError) Errs() []error {
	return l.errors
}
