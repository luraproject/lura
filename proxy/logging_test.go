// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/luraproject/lura/v2/logging"
)

func TestNewLoggingMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := logging.NewLogger("INFO", buff, "pref")
	mw := NewLoggingMiddleware(logger, "supu")
	mw(explosiveProxy(t), explosiveProxy(t))
}

func TestNewLoggingMiddleware_ok(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := logging.NewLogger("DEBUG", buff, "pref")
	resp := &Response{IsComplete: true}
	mw := NewLoggingMiddleware(logger, "supu")
	p := mw(dummyProxy(resp))
	r, err := p(context.Background(), &Request{})
	if r != resp {
		t.Error("The proxy didn't return the expected response")
		return
	}
	if err != nil {
		t.Errorf("The proxy returned an unexpected error: %s", err.Error())
		return
	}
	logMsg := buff.String()
	if strings.Count(logMsg, "pref") != 3 {
		t.Error("The logs don't have the injected prefix")
	}
	if strings.Count(logMsg, "INFO") != 2 {
		t.Error("The logs don't have the expected INFO messages")
	}
	if strings.Count(logMsg, "DEBU") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if !strings.Contains(logMsg, "[SUPU] Calling backend") {
		t.Error("The logs didn't mark the start of the execution")
	}
	if !strings.Contains(logMsg, "[SUPU] Call to backend took") {
		t.Error("The logs didn't mark the end of the execution")
	}
}

func TestNewLoggingMiddleware_erroredResponse(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := logging.NewLogger("DEBUG", buff, "pref")
	resp := &Response{IsComplete: true}
	mw := NewLoggingMiddleware(logger, "supu")
	expextedError := fmt.Errorf("NO-body expects the %s Inquisition!", "Spanish")
	p := mw(func(_ context.Context, _ *Request) (*Response, error) {
		return resp, expextedError
	})
	r, err := p(context.Background(), &Request{})
	if r != resp {
		t.Error("The proxy didn't return the expected response")
		return
	}
	if err != expextedError {
		t.Errorf("The proxy didn't return the expected error: %s", err.Error())
		return
	}
	logMsg := buff.String()
	if strings.Count(logMsg, "pref") != 4 {
		t.Error("The logs don't have the injected prefix")
	}
	if strings.Count(logMsg, "INFO") != 2 {
		t.Error("The logs don't have the expected INFO messages")
	}
	if strings.Count(logMsg, "DEBU") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if strings.Count(logMsg, "WARN") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if !strings.Contains(logMsg, "[SUPU] Call to backend failed: NO-body expects the Spanish Inquisition!") {
		t.Error("The logs didn't mark the fail of the execution")
	}
	if !strings.Contains(logMsg, "[SUPU] Calling backend") {
		t.Error("The logs didn't mark the start of the execution")
	}
	if !strings.Contains(logMsg, "[SUPU] Call to backend took") {
		t.Error("The logs didn't mark the end of the execution")
	}
}

func TestNewLoggingMiddleware_nullResponse(t *testing.T) {
	buff := bytes.NewBuffer(make([]byte, 1024))
	logger, _ := logging.NewLogger("DEBUG", buff, "pref")
	mw := NewLoggingMiddleware(logger, "supu")
	p := mw(dummyProxy(nil))
	r, err := p(context.Background(), &Request{})
	if r != nil {
		t.Error("The proxy didn't return the expected response")
		return
	}
	if err != nil {
		t.Errorf("The proxy returned an unexpected error: %s", err.Error())
		return
	}
	logMsg := buff.String()
	if strings.Count(logMsg, "pref") != 4 {
		t.Error("The logs don't have the injected prefix")
	}
	if strings.Count(logMsg, "INFO") != 2 {
		t.Error("The logs don't have the expected INFO messages")
	}
	if strings.Count(logMsg, "DEBU") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if strings.Count(logMsg, "WARN") != 1 {
		t.Error("The logs don't have the expected DEBUG messages")
	}
	if !strings.Contains(logMsg, "[SUPU] Call to backend returned a null response") {
		t.Error("The logs didn't mark the fail of the execution")
	}
	if !strings.Contains(logMsg, "[SUPU] Calling backend") {
		t.Error("The logs didn't mark the start of the execution")
	}
	if !strings.Contains(logMsg, "[SUPU] Call to backend took") {
		t.Error("The logs didn't mark the end of the execution")
	}
}
