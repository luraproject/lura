package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
)

func TestNewMergeDataMiddleware_ok(t *testing.T) {
	timeout := 500
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(&endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"tupu": true}, IsComplete: true}))
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{})
	if err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if out == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
		if len(out.Data) != 2 {
			t.Errorf("We weren't expecting a partial response but we got %v!\n", out)
		}
		if !out.IsComplete {
			t.Errorf("We were expecting a completed response but we got an incompleted one!\n")
		}
	}
}

func TestNewMergeDataMiddleware_partialTimeout(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(&endpoint)
	p := mw(
		delayedProxy(t, time.Duration(timeout/2)*time.Millisecond, &Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		delayedProxy(t, time.Duration(5*timeout)*time.Millisecond, nil))
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{})
	if err == nil || err.Error() != "context deadline exceeded" {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if out == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
		if len(out.Data) != 1 {
			t.Errorf("We were expecting a partial response but we got %v!\n", out)
		}
		if out.IsComplete {
			t.Errorf("We were expecting an incompleted response but we got a completed one!\n")
		}
	}
}

func TestNewMergeDataMiddleware_partial(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(&endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		dummyProxy(&Response{}))
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{})
	if err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if out == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
		if len(out.Data) != 1 {
			t.Errorf("We were expecting a partial response but we got %v!\n", out)
		}
		if out.IsComplete {
			t.Errorf("We were expecting an incompleted response but we got a completed one!\n")
		}
	}
}

func TestNewMergeDataMiddleware_nullResponse(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
	}
	mw := NewMergeDataMiddleware(&endpoint)

	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := mw(NoopProxy, NoopProxy)(context.Background(), &Request{})
	if err != errNullResult {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if out == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
		if len(out.Data) != 0 {
			t.Errorf("We were expecting a partial response but we got %v!\n", out.Data)
		}
		if out.IsComplete {
			t.Errorf("We were expecting an incompleted response but we got a completed one!\n")
		}
	}
}

func TestNewMergeDataMiddleware_timeout(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(&endpoint)
	p := mw(
		delayedProxy(t, time.Duration(5*timeout)*time.Millisecond, nil),
		delayedProxy(t, time.Duration(5*timeout)*time.Millisecond, nil))
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{})
	if err == nil || err.Error() != "context deadline exceeded" {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if out == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
		if len(out.Data) > 0 {
			t.Errorf("We weren't expecting a response but we got one!\n")
		}
		if out.IsComplete {
			t.Errorf("We were expecting an incompleted response but we got a completed one!\n")
		}
	}
}

func TestNewMergeDataMiddleware_notEnoughBackends(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend},
	}
	mw := NewMergeDataMiddleware(&endpoint)
	mw(explosiveProxy(t), explosiveProxy(t))
}

func TestNewMergeDataMiddleware_notEnoughProxies(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
	}
	mw := NewMergeDataMiddleware(&endpoint)
	mw(NoopProxy)
}

func TestNewMergeDataMiddleware_noBackends(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	endpoint := config.EndpointConfig{}
	NewMergeDataMiddleware(&endpoint)
}
