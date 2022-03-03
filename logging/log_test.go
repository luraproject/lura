// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"os"
	"os/exec"
	"regexp"
	"testing"
)

const (
	debugMsg    = "Debug msg"
	infoMsg     = "Info msg"
	warningMsg  = "Warning msg"
	errorMsg    = "Error msg"
	criticalMsg = "Critical msg"
	fatalMsg    = "Fatal msg"
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
	if err != ErrInvalidLogLevel {
		t.Errorf("The factory didn't return the expected error. Got: %s", err.Error())
	}
}

func TestNewLogger_fatal(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		l, err := NewLogger("Critical", bytes.NewBuffer(make([]byte, 1024)), "pref")
		if err != nil {
			t.Error("The factory returned an expected error:", err.Error())
			return
		}
		l.Fatal("crash!!!")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestNewLogger_fatal")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
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
