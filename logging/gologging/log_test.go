package gologging

import (
	"bytes"
	"regexp"
	"testing"

	gologging "github.com/op/go-logging"
)

const (
	debugMsg    = "Debug msg"
	infoMsg     = "Info msg"
	warningMsg  = "Warning msg"
	errorMsg    = "Error msg"
	criticalMsg = "Critical msg"
)

func TestNewLogger(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
	regexps := []*regexp.Regexp{
		regexp.MustCompile(debugMsg),
		regexp.MustCompile(infoMsg),
		regexp.MustCompile(warningMsg),
		regexp.MustCompile(errorMsg),
		regexp.MustCompile(criticalMsg),
	}

	for i, level := range levels {
		output := logSomeStuff(level)
		for j := i; j < len(regexps); j++ {
			if !regexps[j].MatchString(output) {
				t.Errorf("The output doesn't contain the expected msg for the level: %s. [%s]", level, output)
			}
		}
	}
}

func TestNewLogger_unknownLevel(t *testing.T) {
	_, err := NewLogger("UNKNOWN", bytes.NewBuffer(make([]byte, 1024)), "pref")
	if err == nil {
		t.Error("The factory didn't return the expected error")
		return
	}
	if err != gologging.ErrInvalidLogLevel {
		t.Errorf("The factory didn't return the expected error. Got: %s", err.Error())
	}
}

func logSomeStuff(level string) string {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := NewLogger(level, buff, "pref")

	logger.Debug(debugMsg)
	logger.Info(infoMsg)
	logger.Warning(warningMsg)
	logger.Error(errorMsg)
	logger.Critical(criticalMsg)

	return buff.String()
}
