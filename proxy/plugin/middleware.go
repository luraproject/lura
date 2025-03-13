// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"fmt"

	"github.com/luraproject/lura/v2/logging"
	luraplugin "github.com/luraproject/lura/v2/plugin"
)

// MiddlewareNamespace is the internal namespace for the register to be used with middlewares
const MiddlewareNamespace = "github.com/devopsfaith/krakend/proxy/plugin/middleware"

// MiddlewareFactory is function that, given an extra_config and a middleware returns another middleware with custom logic wrapping the one passed as argument
type MiddlewareFactory func(map[string]interface{}, func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error)

// RegisterModifier registers the injected modifier factory with the given name at the selected namespace
func RegisterMiddleware(
	name string,
	middlewareFactory func(map[string]interface{}, func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error),
) {
	pluginRegister.Register(MiddlewareNamespace, name, middlewareFactory)
}

// MiddlewareRegisterer defines the interface for the plugins to expose in order to be able to be loaded/registered
type MiddlewareRegisterer interface {
	RegisterMiddlewares(func(
		name string,
		middlewareFactory func(map[string]interface{}, func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error),
	))
}

type RegisterMiddlewareFunc func(
	name string,
	middlewareFactory func(map[string]interface{}, func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error),
)

func GetMiddleware(name string) (MiddlewareFactory, bool) {
	r, ok := pluginRegister.Get(MiddlewareNamespace)
	if !ok {
		return nil, ok
	}
	m, ok := r.Get(name)
	if !ok {
		return nil, ok
	}
	res, ok := m.(func(map[string]interface{}, func(context.Context, interface{}) (interface{}, error)) func(context.Context, interface{}) (interface{}, error))
	if !ok {
		return nil, ok
	}
	return MiddlewareFactory(res), ok
}

func LoadMiddlewares(ctx context.Context, path, pattern string, rmf RegisterMiddlewareFunc, logger logging.Logger) (int, error) {
	plugins, err := luraplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}

	var errors []error

	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := openMiddleware(ctx, pluginName, rmf, logger); err != nil {
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

func openMiddleware(ctx context.Context, pluginName string, rmf RegisterMiddlewareFunc, logger logging.Logger) (err error) {
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
	r, err := p.Lookup("MiddlewareRegisterer")
	if err != nil {
		return err
	}
	registerer, ok := r.(MiddlewareRegisterer)
	if !ok {
		return fmt.Errorf("middleware plugin loader: unknown type")
	}

	registerExtraComponents(r, ctx, logger)

	registerer.RegisterMiddlewares(rmf)
	return
}
