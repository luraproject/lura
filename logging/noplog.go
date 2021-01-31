package logging

import (
	"io"
	"os"
)

type NopLogger struct {
	Level  int
	Prefix string
}

// NewNopLogger creates and returns a NopLogger object
func NewNopLogger(level string, out io.Writer, prefix string) (NopLogger, error) {
	return NopLogger{Level: 0, Prefix: prefix}, nil
}

// Debug logs a message using DEBUG as log level.
func (l *NopLogger) Debug(args ...interface{}) {}

// Info logs a message using INFO as log level.
func (l *NopLogger) Info(args ...interface{}) {}

// Warning logs a message using WARNING as log level.
func (l *NopLogger) Warning(args ...interface{}) {}

// Error logs a message using ERROR as log level.
func (l *NopLogger) Error(args ...interface{}) {}

// Critical logs a message using CRITICAL as log level.
func (l *NopLogger) Critical(args ...interface{}) {}

// Fatal is equivalent to l.Critical(fmt.Sprint()) followed by a call to os.Exit(1).
func (l *NopLogger) Fatal(args ...interface{}) {
	os.Exit(1)
}

func (l *NopLogger) prependLog(args ...interface{}) {}
