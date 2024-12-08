package pluginproxy

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"plugin"
	"strings"

	"github.com/luraproject/lura/v2/logging"
	luraplugin "github.com/luraproject/lura/v2/plugin"
	"github.com/luraproject/lura/v2/register"
)

const Namespace = "github.com/devopsfaith/krakend/proxy/pluginproxy"

var proxyRegister = register.New()

type (
	Handler        = func(context.Context, map[string]interface{}, ProxyWrapper) ProxyWrapper
	ProxyWrapper   = func(context.Context, RequestWrapper) (ResponseWrapper, error)
	RequestWrapper = interface {
		Params() map[string]string
		Headers() map[string][]string
		Body() io.ReadCloser
		Method() string
		URL() *url.URL
		Query() url.Values
		Path() string
	}
	ResponseWrapper = interface {
		Data() map[string]interface{}
		Io() io.Reader
		IsComplete() bool
		Headers() map[string][]string
		StatusCode() int
	}
)

func RegisterProxies(name string, handler Handler) {
	proxyRegister.Register(Namespace, name, handler)
}

type ProxyRegisterer interface {
	RegisterProxies(func(name string, handler Handler))
}

type LoggerRegisterer interface {
	RegisterLogger(interface{})
}

type ContextRegisterer interface {
	RegisterContext(context.Context)
}

type RegisterProxyFunc func(
	name string,
	handler Handler,
)

func Load(path, pattern string, rpf RegisterProxyFunc) (int, error) {
	return LoadWithLogger(path, pattern, rpf, nil)
}

func LoadWithLogger(path, pattern string, rpf RegisterProxyFunc, logger logging.Logger) (int, error) {
	plugins, err := luraplugin.Scan(path, pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, rpf, logger)
}

func load(plugins []string, rpf RegisterProxyFunc, logger logging.Logger) (int, error) {
	errors := []error{}
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, rpf, logger); err != nil {
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

func open(pluginName string, rpf RegisterProxyFunc, logger logging.Logger) (err error) {
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
	r, err = p.Lookup("ProxyRegisterer")
	if err != nil {
		return
	}
	registerer, ok := r.(ProxyRegisterer)
	if !ok {
		return fmt.Errorf("proxy plugin loader: unknown type")
	}

	if logger != nil {
		if lr, ok := r.(LoggerRegisterer); ok {
			lr.RegisterLogger(logger)
		}
	}

	registerer.RegisterProxies(rpf)
	return
}

func GetProxy(name string) (Handler, bool) {
	r, ok := proxyRegister.Get(Namespace)
	if !ok {
		return nil, ok
	}
	p, ok := r.Get(name)
	if !ok {
		return nil, ok
	}
	res, ok := p.(Handler)
	if !ok {
		return nil, ok
	}
	return res, ok
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
