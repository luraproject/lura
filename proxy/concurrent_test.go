// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/config"
)

func TestNewConcurrentMiddleware_ok(t *testing.T) {
	timeout := 700
	totalCalls := 3
	backend := config.Backend{
		ConcurrentCalls: totalCalls,
		Timeout:         time.Duration(timeout) * time.Millisecond,
	}
	expected := Response{
		Data:       map[string]interface{}{"supu": 42, "tupu": true, "foo": "bar"},
		IsComplete: true,
	}
	mw := NewConcurrentMiddleware(&backend)
	mustEnd := time.After(time.Duration(timeout) * time.Millisecond)
	result, err := mw(dummyProxy(&expected))(context.Background(), &Request{})
	if err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
	}
	if result == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	if !result.IsComplete {
		t.Errorf("The proxy returned an incomplete result: %v\n", result)
	}
	if v, ok := result.Data["supu"]; !ok || v.(int) != 42 {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
	if v, ok := result.Data["tupu"]; !ok || !v.(bool) {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
	if v, ok := result.Data["foo"]; !ok || v.(string) != "bar" {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
}

func TestNewConcurrentMiddleware_okAfterKo(t *testing.T) {
	timeout := 700
	totalCalls := 3
	backend := config.Backend{
		ConcurrentCalls: totalCalls,
		Timeout:         time.Duration(timeout) * time.Millisecond,
	}
	expected := Response{
		Data:       map[string]interface{}{"supu": 42, "tupu": true, "foo": "bar"},
		IsComplete: true,
	}
	mw := NewConcurrentMiddleware(&backend)

	calls := uint64(0)
	mock := func(_ context.Context, _ *Request) (*Response, error) {
		total := atomic.AddUint64(&calls, 1)
		if total%2 == 0 {
			return &expected, nil
		}
		return nil, nil
	}
	mustEnd := time.After(time.Duration(timeout) * time.Millisecond)
	result, err := mw(mock)(context.Background(), &Request{})

	if result == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	if err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
	}
	if !result.IsComplete {
		t.Errorf("The proxy returned an incomplete result: %v\n", result)
	}
	if v, ok := result.Data["supu"]; !ok || v.(int) != 42 {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
	if v, ok := result.Data["tupu"]; !ok || !v.(bool) {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
	if v, ok := result.Data["foo"]; !ok || v.(string) != "bar" {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
}

func TestNewConcurrentMiddleware_timeout(t *testing.T) {
	timeout := 100
	totalCalls := 3
	backend := config.Backend{
		ConcurrentCalls: totalCalls,
		Timeout:         time.Duration(timeout) * time.Millisecond,
	}
	mw := NewConcurrentMiddleware(&backend)
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)

	response, err := mw(delayedProxy(t, time.Duration(5*timeout)*time.Millisecond, &Response{}))(context.Background(), &Request{})
	if err == nil || err.Error() != "context deadline exceeded" {
		t.Errorf("The middleware didn't propagate a timeout error: %s\n", err)
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one: %v\n", response)
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response at this point in time!\n")
		return
	default:
	}
}

func TestNewConcurrentMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	backend := config.Backend{ConcurrentCalls: 2}
	mw := NewConcurrentMiddleware(&backend)
	mw(explosiveProxy(t), explosiveProxy(t))
}

func TestNewConcurrentMiddleware_insuficientConcurrentCalls(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	backend := config.Backend{ConcurrentCalls: 1}
	NewConcurrentMiddleware(&backend)
}
