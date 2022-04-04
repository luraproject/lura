// SPDX-License-Identifier: Apache-2.0

/*
	Package logging provides a simple logger interface and implementations
*/
package logging

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Logger collects logging information at several levels
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}

const (
	// LEVEL_DEBUG = 0
	LEVEL_DEBUG = iota
	// LEVEL_INFO = 1
	LEVEL_INFO
	// LEVEL_WARNING = 2
	LEVEL_WARNING
	// LEVEL_ERROR = 3
	LEVEL_ERROR
	// LEVEL_CRITICAL = 4
	LEVEL_CRITICAL
)

var (
	// ErrInvalidLogLevel is used when an invalid log level has been used.
	ErrInvalidLogLevel = errors.New("invalid log level")
	defaultLogger      = BasicLogger{Level: LEVEL_CRITICAL, Logger: log.New(os.Stderr, "", log.LstdFlags)}
	logLevels          = map[string]int{
		"DEBUG":    LEVEL_DEBUG,
		"INFO":     LEVEL_INFO,
		"WARNING":  LEVEL_WARNING,
		"ERROR":    LEVEL_ERROR,
		"CRITICAL": LEVEL_CRITICAL,
	}
	// NoOp is the NO-OP logger
	NoOp, _ = NewLogger("CRITICAL", ioutil.Discard, "")
)

// NewLogger creates and returns a Logger object
func NewLogger(level string, out io.Writer, prefix string) (BasicLogger, error) {
	l, ok := logLevels[strings.ToUpper(level)]
	if !ok {
		return defaultLogger, ErrInvalidLogLevel
	}
	return BasicLogger{Level: l, Prefix: prefix, Logger: log.New(out, "", log.LstdFlags)}, nil
}

type BasicLogger struct {
	Level  int
	Prefix string
	Logger *log.Logger
}

// Debug logs a message using DEBUG as log level.
func (l BasicLogger) Debug(v ...interface{}) {
	if l.Level > LEVEL_DEBUG {
		return
	}
	l.prependLog("DEBUG:", v...)
}

// Info logs a message using INFO as log level.
func (l BasicLogger) Info(v ...interface{}) {
	if l.Level > LEVEL_INFO {
		return
	}
	l.prependLog("INFO:", v...)
}

// Warning logs a message using WARNING as log level.
func (l BasicLogger) Warning(v ...interface{}) {
	if l.Level > LEVEL_WARNING {
		return
	}
	l.prependLog("WARNING:", v...)
}

// Error logs a message using ERROR as log level.
func (l BasicLogger) Error(v ...interface{}) {
	if l.Level > LEVEL_ERROR {
		return
	}
	l.prependLog("ERROR:", v...)
}

// Critical logs a message using CRITICAL as log level.
func (l BasicLogger) Critical(v ...interface{}) {
	l.prependLog("CRITICAL:", v...)
}

// Fatal is equivalent to l.Critical(fmt.Sprint()) followed by a call to os.Exit(1).
func (l BasicLogger) Fatal(v ...interface{}) {
	l.prependLog("FATAL:", v...)
	os.Exit(1)
}

func (l BasicLogger) prependLog(level string, v ...interface{}) {
	msg := make([]interface{}, len(v)+2)
	msg[0] = l.Prefix
	msg[1] = level
	copy(msg[2:], v)
	l.Logger.Println(msg...)
}
