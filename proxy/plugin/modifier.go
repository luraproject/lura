// SPDX-License-Identifier: Apache-2.0

/*
Package plugin provides tools for loading and registering proxy plugins
*/
package plugin

import (
	"context"
	"fmt"

	"github.com/luraproject/lura/v2/logging"
	luraplugin "github.com/luraproject/lura/v2/plugin"
	"github.com/luraproject/lura/v2/register"
)

const (
	// requestNamespace is the internal namespace for the register to be used with request modifiers
	requestNamespace = "github.com/devopsfaith/krakend/proxy/plugin/request"
	// responseNamespace is the internal namespace for the register to be used with response modifiers
	responseNamespace = "github.com/devopsfaith/krakend/proxy/plugin/response"
)

var pluginRegister = register.New()

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
	r, ok := pluginRegister.Get(namespace)
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
		pluginRegister.Register(requestNamespace, name, modifierFactory)
	}
	if appliesToResponse {
		pluginRegister.Register(responseNamespace, name, modifierFactory)
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

type ContextRegisterer interface {
	RegisterContext(context.Context)
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
	return LoadWithLoggerAndContext(context.Background(), path, pattern, rmf, logger)
}

func LoadWithLoggerAndContext(ctx context.Context, path, pattern string, rmf RegisterModifierFunc, logger logging.Logger) (int, error) {
	plugins, err := luraplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}

	var errors []error

	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := openModifier(ctx, pluginName, rmf, logger); err != nil {
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

func openModifier(ctx context.Context, pluginName string, rmf RegisterModifierFunc, logger logging.Logger) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	p, err := pluginOpener(pluginName)
	if err != nil {
		return err
	}
	r, err := p.Lookup("ModifierRegisterer")
	if err != nil {
		return err
	}
	registerer, ok := r.(Registerer)
	if !ok {
		return fmt.Errorf("modifier plugin loader: unknown type")
	}

	registerExtraComponents(r, ctx, logger)

	registerer.RegisterModifiers(rmf)
	return
}
