package proxy

import (
	"context"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestNewMergeDataMiddleware_simpleFiltering(t *testing.T) {
	timeout := 500
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	RegisterBackendFiltererFactory(func(_ *config.EndpointConfig) ([]BackendFilterer, error) {
		return []BackendFilterer{
			func(_ *Request) bool {
				return true
			},
			func(r *Request) bool {
				return r.Headers["X-Filter"][0] == "supu"
			},
		}, nil
	})

	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"tupu": true}, IsComplete: true}))
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{Headers: map[string][]string{"X-Filter": {"meh"}}})
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
		if len(out.Data) != 1 || out.Data["supu"] != 42 {
			t.Errorf("We were expecting a response from just a backend, but we got %v!\n", out)
		}
		if !out.IsComplete {
			t.Errorf("We were expecting a completed response but we got an incompleted one!\n")
		}
	}
}

func TestNewMergeDataMiddleware_sequentialFiltering(t *testing.T) {
	timeout := 1000
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{
			{URLPattern: "/"},
			{URLPattern: "/aaa/{{.Resp0_string}}"},
			{URLPattern: "/hit-me/{{.Resp1_tupu}}"},
		},
		Timeout: time.Duration(timeout) * time.Millisecond,
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				isSequentialKey: true,
			},
		},
	}

	RegisterBackendFiltererFactory(func(_ *config.EndpointConfig) ([]BackendFilterer, error) {
		return []BackendFilterer{
			func(_ *Request) bool {
				return true
			},
			func(r *Request) bool {
				return false
			},
			func(_ *Request) bool {
				return true
			},
		}, nil
	})

	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{
			"int":    42,
			"string": "some",
			"bool":   true,
			"float":  3.14,
			"struct": map[string]interface{}{
				"foo": "bar",
				"struct": map[string]interface{}{
					"foo": "bar",
					"struct": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
			"array":      []interface{}{"1", "2"},
			"propagated": "everywhere",
		}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"tupu": "foo"}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"final": "meh"}, IsComplete: true}),
	)
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{
		Params: map[string]string{},
	})
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
		if len(out.Data) != 8 {
			t.Errorf("We were expecting a response from just two backends, but we got %v!\n", out)
		}
		if !out.IsComplete {
			t.Errorf("We were expecting a completed response but we got an incompleted one!\n")
		}
	}
}
