// SPDX-License-Identifier: Apache-2.0

package logging

// AnnotatedLogger wraps a Logger and appends information to
// each one of the received logs.
type AnnotatedLogger struct {
	appendAnnotation string
	wrapped          Logger
}

// NewAnnotatedLogger creates and returns a Logger object
func NewAnnotatedLogger(l Logger, appendAnnotation string) (AnnotatedLogger, error) {
	if al, ok := l.(AnnotatedLogger); ok {
		// if the given logger is already an annotated logger, we
		// can combine the two appended anotations and save a wrap around
		if len(al.appendAnnotation) > 0 {
			appendAnnotation = al.appendAnnotation + " " + appendAnnotation
		}
		l = al.wrapped
	}
	return AnnotatedLogger{
		appendAnnotation: appendAnnotation,
		wrapped:          l,
	}, nil
}

// Debug logs a message using DEBUG as log level.
func (l AnnotatedLogger) Debug(v ...interface{}) {
	l.wrapped.Debug(l.appendLog(v))
}

// Info logs a message using INFO as log level.
func (l AnnotatedLogger) Info(v ...interface{}) {
	l.wrapped.Info(l.appendLog(v))
}

// Warning logs a message using WARNING as log level.
func (l AnnotatedLogger) Warning(v ...interface{}) {
	l.wrapped.Warning(l.appendLog(v))
}

// Error logs a message using ERROR as log level.
func (l AnnotatedLogger) Error(v ...interface{}) {
	l.wrapped.Error(l.appendLog(v))
}

// Critical logs a message using CRITICAL as log level.
func (l AnnotatedLogger) Critical(v ...interface{}) {
	l.wrapped.Critical(l.appendLog(v))
}

// Fatal is equivalent to l.Critical(fmt.Sprint()) followed by a call to os.Exit(1).
func (l AnnotatedLogger) Fatal(v ...interface{}) {
	l.wrapped.Fatal(l.appendLog(v))
}

func (l AnnotatedLogger) appendLog(v ...interface{}) []interface{} {
	if len(l.appendAnnotation) == 0 {
		return v
	}
	msg := make([]interface{}, len(v)+1)
	copy(msg, v)
	msg[len(v)] = l.appendAnnotation
	return msg
}
