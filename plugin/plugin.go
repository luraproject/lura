package plugin

import (
	"fmt"
	"io/ioutil"
	"plugin"
	"strings"
)

// Load loads all the plugins in pluginFolder with pattern in its filename.
// It returns the number of plugins loaded and an error if something goes wrong.
func Load(pluginFolder, pattern string) (int, error) {
	plugins, err := scan(pluginFolder, pattern)
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
		if _, err := plugin.Open(pluginName); err != nil {
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

type loaderError struct {
	errors []error
}

func (l loaderError) Error() string {
	msgs := make([]string, len(l.errors))
	for i, err := range l.errors {
		msgs[i] = err.Error()
	}
	return fmt.Sprintf("plugin loader found %d error(s): \n%s", len(msgs), strings.Join(msgs, "\n"))
}
