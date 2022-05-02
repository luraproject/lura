// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"net/http"
)

func main() {}

var ModifierRegisterer = registerer("lura-error-example")

var logger Logger = nil

type registerer string

func (r registerer) RegisterModifiers(f func(
	name string,
	modifierFactory func(map[string]interface{}) func(interface{}) (interface{}, error),
	appliesToRequest bool,
	appliesToResponse bool,
)) {
	f(string(r)+"-request", r.requestModifierFactory, true, false)
	f(string(r)+"-response", r.reqsponseModifierFactory, false, true)
}

func (registerer) RegisterLogger(in interface{}) {
	l, ok := in.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ModifierRegisterer))
}

func (registerer) requestModifierFactory(_ map[string]interface{}) func(interface{}) (interface{}, error) {
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Request modifier injected", ModifierRegisterer))
	return func(_ interface{}) (interface{}, error) {
		logger.Debug(fmt.Sprintf("[PLUGIN: %s] Rejecting request", ModifierRegisterer))
		return nil, requestErr
	}
}

func (registerer) reqsponseModifierFactory(_ map[string]interface{}) func(interface{}) (interface{}, error) {
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Response modifier injected", ModifierRegisterer))
	return func(_ interface{}) (interface{}, error) {
		logger.Debug(fmt.Sprintf("[PLUGIN: %s] Replacing response", ModifierRegisterer))
		return nil, responseErr
	}
}

type customError struct {
	error
	statusCode int
}

func (r customError) StatusCode() int { return r.statusCode }

var (
	requestErr = customError{
		error:      errors.New("request rejected just because"),
		statusCode: http.StatusTeapot,
	}
	responseErr = customError{
		error:      errors.New("response replaced because reasons"),
		statusCode: http.StatusTeapot,
	}
)

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}
