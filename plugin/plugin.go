package plugin

import (
	"fmt"
	"io/ioutil"
	"plugin"
	"strings"

	"github.com/devopsfaith/krakend/config"
)

// Load loads all the plugins in pluginFolder with pattern in its filename.
// It returns the number of plugins loaded and an error if something goes wrong.
func Load(cfg config.Plugin) (int, error) {
	plugins, err := scan(cfg.Folder, cfg.Pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins)
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

func load(plugins []string) (int, error) {
	errors := []error{}
	loadedPlugins := 0
	for k, pluginName := range plugins {
		if err := open(pluginName); err != nil {
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

func open(pluginName string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	_, err = pluginOpener(pluginName)
	return
}

// pluginOpener keeps the plugin open function in a var for easy testing
var pluginOpener = plugin.Open

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
