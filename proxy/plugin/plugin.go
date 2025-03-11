// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"fmt"
	"plugin"
	"strings"

	"github.com/luraproject/lura/v2/logging"
)

// Namespace is the namespace for the extra_config section
const Namespace = "github.com/devopsfaith/krakend/proxy/plugin"

// Plugin is the interface of the loaded plugins
type Plugin interface {
	Lookup(name string) (plugin.Symbol, error)
}

// pluginOpener keeps the plugin open function in a var for easy testing
var pluginOpener = defaultPluginOpener

func defaultPluginOpener(name string) (Plugin, error) {
	return plugin.Open(name)
}

func registerExtraComponents(r plugin.Symbol, ctx context.Context, l logging.Logger) {
	if l != nil {
		if lr, ok := r.(LoggerRegisterer); ok {
			lr.RegisterLogger(l)
		}
	}

	if lr, ok := r.(ContextRegisterer); ok {
		lr.RegisterContext(ctx)
	}
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
