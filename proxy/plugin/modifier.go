// SPDX-License-Identifier: Apache-2.0

/*
	Package plugin provides tools for loading and registering proxy plugins
*/
package plugin

import (
	"fmt"
	"plugin"
	"strings"

	"github.com/luraproject/lura/v2/logging"
	luraplugin "github.com/luraproject/lura/v2/plugin"
	"github.com/luraproject/lura/v2/register"
)

const (
	// Namespace is the namespace for the extra_config section
	Namespace = "github.com/devopsfaith/krakend/proxy/plugin"
	// requestNamespace is the internal namespace for the register to be used with request modifiers
	requestNamespace = "github.com/devopsfaith/krakend/proxy/plugin/request"
	// responseNamespace is the internal namespace for the register to be used with response modifiers
	responseNamespace = "github.com/devopsfaith/krakend/proxy/plugin/response"
)

var modifierRegister = register.New()

// ModifierFactory is a function that, given a config passed as a map, returns a modifier
type ModifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error)

// GetRequestModifier returns a ModifierFactory from the request namespace by name
func GetRequestModifier(name string) (ModifierFactory, bool) {
	return getModifier(requestNamespace, name)
}

// GetResponseModifier returns a ModifierFactory from the response namespace by name
func GetResponseModifier(name string) (ModifierFactory, bool) {
	return getModifier(responseNamespace, name)
}

func getModifier(namespace, name string) (ModifierFactory, bool) {
	r, ok := modifierRegister.Get(namespace)
	if !ok {
		return nil, ok
	}
	m, ok := r.Get(name)
	if !ok {
		return nil, ok
	}
	res, ok := m.(func(map[string]interface{}) func(interface{}) (interface{}, error))
	if !ok {
		return nil, ok
	}
	return ModifierFactory(res), ok
}

// RegisterModifier registers the injected modifier factory with the given name at the selected namespace
func RegisterModifier(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
) {
	if appliesToRequest {
		modifierRegister.Register(requestNamespace, name, modifierFactory)
	}
	if appliesToResponse {
		modifierRegister.Register(responseNamespace, name, modifierFactory)
	}
}

// Registerer defines the interface for the plugins to expose in order to be able to be loaded/registered
type Registerer interface {
	RegisterModifiers(func(
		name string,
		modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
		appliesToRequest bool,
		appliesToResponse bool,
	))
}

type LoggerRegisterer interface {
	RegisterLogger(interface{})
}

// RegisterModifierFunc type is the function passed to the loaded Registerers
type RegisterModifierFunc func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)

// Load scans the given path using the pattern and registers all the found modifier plugins into the rmf
func Load(path, pattern string, rmf RegisterModifierFunc) (int, error) {
	return LoadWithLogger(path, pattern, rmf, nil)
}

// LoadWithLogger scans the given path using the pattern and registers all the found modifier plugins into the rmf
func LoadWithLogger(path, pattern string, rmf RegisterModifierFunc, logger logging.Logger) (int, error) {
	plugins, err := luraplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, rmf, logger)
}

func load(plugins []string, rmf RegisterModifierFunc, logger logging.Logger) (int, error) {
	errors := []error{}
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, rmf, logger); err != nil {
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

func open(pluginName string, rmf RegisterModifierFunc, logger logging.Logger) (err error) {
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
	r, err = p.Lookup("ModifierRegisterer")
	if err != nil {
		return
	}
	registerer, ok := r.(Registerer)
	if !ok {
		return fmt.Errorf("modifier plugin loader: unknown type")
	}

	if logger != nil {
		if lr, ok := r.(LoggerRegisterer); ok {
			lr.RegisterLogger(logger)
		}
	}

	registerer.RegisterModifiers(rmf)
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
