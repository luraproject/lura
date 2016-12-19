//Package gologging provides a logger implementation based on the github.com/op/go-logging pkg
package gologging

import (
	"io"

	gologging "github.com/op/go-logging"

	"github.com/devopsfaith/krakend/logging"
)

// NewLogger returns a krakend logger wrapping a gologging logger
func NewLogger(level string, out io.Writer, prefix string) (logging.Logger, error) {
	module := "KRAKEND"
	log := gologging.MustGetLogger(module)
	logBackend := gologging.NewLogBackend(out, prefix, 0)
	format := gologging.MustStringFormatter(
		` %{time:2006/01/02 - 15:04:05.000} %{color}▶ %{level:.4s}%{color:reset} %{message}`,
	)
	backendFormatter := gologging.NewBackendFormatter(logBackend, format)
	backendLeveled := gologging.AddModuleLevel(backendFormatter)
	logLevel, err := gologging.LogLevel(level)
	if err != nil {
		return nil, err
	}
	backendLeveled.SetLevel(logLevel, module)
	gologging.SetBackend(backendLeveled)
	return Logger{log}, nil
}

// Logger is a wrapper over a github.com/op/go-logging logger
type Logger struct {
	Logger *gologging.Logger
}

// Debug implements the logger interface
func (l Logger) Debug(v ...interface{}) {
	l.Logger.Debug(v...)
}

// Info implements the logger interface
func (l Logger) Info(v ...interface{}) {
	l.Logger.Info(v...)
}

// Warning implements the logger interface
func (l Logger) Warning(v ...interface{}) {
	l.Logger.Warning(v...)
}

// Error implements the logger interface
func (l Logger) Error(v ...interface{}) {
	l.Logger.Error(v...)
}

// Critical implements the logger interface
func (l Logger) Critical(v ...interface{}) {
	l.Logger.Critical(v...)
}

// Fatal implements the logger interface
func (l Logger) Fatal(v ...interface{}) {
	l.Logger.Fatal(v...)
}
