// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestNewMergeDataMiddleware(t *testing.T) {
	tests := []func(t *testing.T){
		testNewMergeDataMiddleware_simpleFiltering,
		testNewMergeDataMiddleware_sequentialFiltering,
		testNewMergeDataMiddleware_empty,
		testNewMergeDataMiddleware_ok,
		testNewMergeDataMiddleware_sequential,
		testNewMergeDataMiddleware_sequential_unavailableParams,
		testNewMergeDataMiddleware_sequential_erroredBackend,
		testNewMergeDataMiddleware_sequential_erroredFirstBackend,
		testNewMergeDataMiddleware_mergeIncompleteResults,
		testNewMergeDataMiddleware_mergeEmptyResults,
		testNewMergeDataMiddleware_partialTimeout,
		testNewMergeDataMiddleware_partial,
		testNewMergeDataMiddleware_nullResponse,
		testNewMergeDataMiddleware_timeout,
		testRegisterResponseCombiner,
	}
	for _, test := range tests {
		ResetBackendFiltererFactory()
		test(t)
	}
}

func testNewMergeDataMiddleware_empty(t *testing.T) {
	timeout := 500 * time.Millisecond
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: timeout,
	}

	expectedErr := errors.New("wait for me")

	erroredProxy := func(_ context.Context, _ *Request) (*Response, error) {
		return nil, expectedErr
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(erroredProxy, erroredProxy)

	mustEnd := time.After(2 * timeout)
	out, err := p(context.Background(), &Request{})
	mErr, ok := err.(mergeError)
	if !ok {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err)
		return
	}
	if len(mErr.errs) != 2 {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err)
		return
	}
	if mErr.errs[0] != mErr.errs[1] || mErr.errs[0] != expectedErr {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err)
		return
	}
	if out != nil {
		t.Errorf("The proxy returned a result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
	}

}

func testNewMergeDataMiddleware_ok(t *testing.T) {
	timeout := 500
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
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

func testNewMergeDataMiddleware_sequential(t *testing.T) {
	timeout := 1000
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{
			{URLPattern: "/"},
			{URLPattern: "/aaa/{{.Resp0_array}}"},
			{URLPattern: "/aaa/{{.Resp0_int}}/{{.Resp0_string}}/{{.Resp0_bool}}/{{.Resp0_float}}/{{.Resp0_struct.foo}}"},
			{URLPattern: "/aaa/{{.Resp0_int}}/{{.Resp0_string}}/{{.Resp0_bool}}/{{.Resp0_float}}/{{.Resp0_struct.foo}}?x={{.Resp1_tupu}}"},
			{URLPattern: "/aaa/{{.Resp0_struct.foo}}/{{.Resp0_struct.struct.foo}}/{{.Resp0_struct.struct.struct.foo}}"},
			{URLPattern: "/zzz", Encoding: "no-op"},
			{URLPattern: "/hit-me"},
		},
		Timeout: time.Duration(timeout) * time.Millisecond,
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				isSequentialKey:        true,
				sequentialPropagateKey: []interface{}{"resp0_propagated", "resp5"},
			},
		},
	}

	expectedBody := "foo"
	checkBody := func(t *testing.T, r *Request) {
		if r.Body == nil {
			t.Error("empty body")
			return
		}
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}

		r.Body.Close()

		if string(b) != expectedBody {
			t.Errorf("unexpected body '%s'", string(b))
		}
	}
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
		func(_ context.Context, r *Request) (*Response, error) {
			checkBody(t, r)
			checkRequestParam(t, r, "Resp0_array", "1,2")
			checkRequestParam(t, r, "Resp0_propagated", "everywhere")
			return &Response{Data: map[string]interface{}{"tupu": "foo"}, IsComplete: true}, nil
		},
		func(_ context.Context, r *Request) (*Response, error) {
			checkBody(t, r)
			checkRequestParam(t, r, "Resp0_int", "42")
			checkRequestParam(t, r, "Resp0_string", "some")
			checkRequestParam(t, r, "Resp0_float", "3.14E+00")
			checkRequestParam(t, r, "Resp0_bool", "true")
			checkRequestParam(t, r, "Resp0_struct.foo", "bar")
			checkRequestParam(t, r, "Resp0_propagated", "everywhere")
			return &Response{Data: map[string]interface{}{"tupu": "foo"}, IsComplete: true}, nil
		},
		func(_ context.Context, r *Request) (*Response, error) {
			checkBody(t, r)
			checkRequestParam(t, r, "Resp0_int", "42")
			checkRequestParam(t, r, "Resp0_string", "some")
			checkRequestParam(t, r, "Resp0_float", "3.14E+00")
			checkRequestParam(t, r, "Resp0_bool", "true")
			checkRequestParam(t, r, "Resp0_struct.foo", "bar")
			checkRequestParam(t, r, "Resp1_tupu", "foo")
			checkRequestParam(t, r, "Resp0_propagated", "everywhere")
			return &Response{Data: map[string]interface{}{"aaaa": []int{1, 2, 3}}, IsComplete: true}, nil
		},
		func(_ context.Context, r *Request) (*Response, error) {
			checkBody(t, r)
			checkRequestParam(t, r, "Resp0_struct.foo", "bar")
			checkRequestParam(t, r, "Resp0_struct.struct.foo", "bar")
			checkRequestParam(t, r, "Resp0_struct.struct.struct.foo", "bar")
			checkRequestParam(t, r, "Resp0_propagated", "everywhere")
			return &Response{Data: map[string]interface{}{"bbbb": []bool{true, false}}, IsComplete: true}, nil
		},
		func(_ context.Context, r *Request) (*Response, error) {
			checkBody(t, r)
			checkRequestParam(t, r, "Resp0_propagated", "everywhere")
			return &Response{Data: map[string]interface{}{}, Io: io.NopCloser(strings.NewReader("hello")), IsComplete: true}, nil
		},
		func(_ context.Context, r *Request) (*Response, error) {
			checkBody(t, r)
			checkRequestParam(t, r, "Resp0_propagated", "everywhere")
			checkRequestParam(t, r, "Resp5", "hello")
			return &Response{Data: map[string]interface{}{}, IsComplete: true}, nil
		},
	)
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{
		Params: map[string]string{},
		Body:   io.NopCloser(strings.NewReader(expectedBody)),
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
		if len(out.Data) != 10 {
			t.Errorf("We weren't expecting a partial response but we got %v!\n", out)
		}
		if !out.IsComplete {
			t.Errorf("We were expecting a completed response but we got an incompleted one!\n")
		}
	}
}

func checkRequestParam(t *testing.T, r *Request, k, v string) {
	if r.Params[k] != v {
		t.Errorf("request without the expected set of params: %s - %+v", k, r.Params)
	}
}

func testNewMergeDataMiddleware_sequential_unavailableParams(t *testing.T) {
	timeout := 500
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{
			{URLPattern: "/"},
			{URLPattern: "/aaa/{{.Resp2_supu}"},
			{URLPattern: "/aaa/{{.Resp0_tupu}}?x={{.Resp1_tupu}}"},
		},
		Timeout: time.Duration(timeout) * time.Millisecond,
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				isSequentialKey: true,
			},
		},
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		func(_ context.Context, r *Request) (*Response, error) {
			if v, ok := r.Params["Resp0_supu"]; ok || v != "" {
				t.Errorf("request with unexpected set of params")
			}
			return &Response{Data: map[string]interface{}{"tupu": "foo"}, IsComplete: true}, nil
		},
		func(_ context.Context, r *Request) (*Response, error) {
			if v, ok := r.Params["Resp0_supu"]; ok || v != "" {
				t.Errorf("request with unexpected set of params")
			}
			if r.Params["Respo_tupu"] != "" {
				t.Errorf("request without the expected set of params")
			}
			if r.Params["Resp1_tupu"] != "foo" {
				t.Errorf("request without the expected set of params")
			}
			return &Response{Data: map[string]interface{}{"aaaa": []int{1, 2, 3}}, IsComplete: true}, nil
		},
	)
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{Params: map[string]string{}})
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
		if len(out.Data) != 3 {
			t.Errorf("We weren't expecting a partial response but we got %v!\n", out)
		}
		if !out.IsComplete {
			t.Errorf("We were expecting a completed response but we got an incompleted one!\n")
		}
	}
}

func testNewMergeDataMiddleware_sequential_erroredBackend(t *testing.T) {
	timeout := 500
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{
			{URLPattern: "/"},
			{URLPattern: "/aaa/{{.Resp0_supu}}"},
			{URLPattern: "/aaa/{{.Resp0_supu}}?x={{.Resp1_tupu}}"},
		},
		Timeout: time.Duration(timeout) * time.Millisecond,
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				isSequentialKey: true,
			},
		},
	}
	expecterErr := errors.New("wait for me")
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		func(_ context.Context, r *Request) (*Response, error) {
			if r.Params["Resp0_supu"] != "42" {
				t.Errorf("request without the expected set of params")
			}
			return nil, expecterErr
		},
		func(_ context.Context, _ *Request) (*Response, error) {
			return nil, nil
		},
	)
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{Params: map[string]string{}})
	if err == nil {
		t.Errorf("The middleware did not propagate an error\n")
		return
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
			t.Errorf("We weren't expecting a partial response but we got %v!\n", out)
		}
		if out.IsComplete {
			t.Errorf("We were not expecting a completed response!\n")
		}
	}
}

func testNewMergeDataMiddleware_sequential_erroredFirstBackend(t *testing.T) {
	timeout := 500
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{
			{URLPattern: "/"},
			{URLPattern: "/aaa/{{.Resp0_supu}}"},
			{URLPattern: "/aaa/{{.Resp0_supu}}?x={{.Resp1_tupu}}"},
		},
		Timeout: time.Duration(timeout) * time.Millisecond,
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				isSequentialKey: true,
			},
		},
	}
	expecterErr := errors.New("wait for me")
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		func(_ context.Context, _ *Request) (*Response, error) {
			return nil, expecterErr
		},
		func(_ context.Context, _ *Request) (*Response, error) {
			t.Error("this backend should never be called")
			return nil, nil
		},
		func(_ context.Context, _ *Request) (*Response, error) {
			t.Error("this backend should never be called")
			return nil, nil
		},
	)
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{Params: map[string]string{}})
	if err != expecterErr {
		t.Errorf("The middleware did not propagate the expected error: %v\n", err)
		return
	}
	if out != nil {
		t.Errorf("The proxy returned a not null result %v", out)
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
	}
}

func testNewMergeDataMiddleware_mergeIncompleteResults(t *testing.T) {
	timeout := 500
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}),
		dummyProxy(&Response{Data: map[string]interface{}{"tupu": true}, IsComplete: false}))
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
			t.Errorf("We were expecting incomplete results merged but we got %v!\n", out.Data)
		}
		if out.IsComplete {
			t.Errorf("We were expecting an incomplete response but we got a completed one!\n")
		}
	}
}

func testNewMergeDataMiddleware_mergeEmptyResults(t *testing.T) {
	timeout := 500
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		dummyProxy(&Response{Data: nil, IsComplete: false}),
		dummyProxy(&Response{Data: nil, IsComplete: false}))
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
		if len(out.Data) != 0 {
			t.Errorf("We were expecting empty data but we got %v!\n", out)
		}
		if out.IsComplete {
			t.Errorf("We were expecting an incomplete response but we got an incompleted one!\n")
		}
	}
}

func testNewMergeDataMiddleware_partialTimeout(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
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

func testNewMergeDataMiddleware_partial(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
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

func testNewMergeDataMiddleware_nullResponse(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)

	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := mw(NoopProxy, NoopProxy)(context.Background(), &Request{})
	if err == nil {
		t.Errorf("The middleware did not propagate the expected error")
	}
	switch mergeErr := err.(type) {
	case mergeError:
		if len(mergeErr.errs) != 2 {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
		if mergeErr.errs[0] != mergeErr.errs[1] {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
		if mergeErr.errs[0] != errNullResult {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
	default:
		t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
	}
	if out != nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
	}
}

func testNewMergeDataMiddleware_timeout(t *testing.T) {
	timeout := 100
	backend := config.Backend{Timeout: time.Duration(timeout) * time.Millisecond}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		delayedProxy(t, time.Duration(5*timeout)*time.Millisecond, nil),
		delayedProxy(t, time.Duration(5*timeout)*time.Millisecond, nil))
	mustEnd := time.After(time.Duration(2*timeout) * time.Millisecond)
	out, err := p(context.Background(), &Request{})
	if err == nil {
		t.Errorf("The middleware did not propagate the expected error")
	}
	switch mergeErr := err.(type) {
	case mergeError:
		if len(mergeErr.errs) != 2 {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
		if mergeErr.errs[0].Error() != mergeErr.errs[1].Error() {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
		if mergeErr.errs[0].Error() != "context deadline exceeded" {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
	default:
		t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
	}
	if out != nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response but we got none\n")
	default:
	}
}

func testRegisterResponseCombiner(t *testing.T) {
	subject := "test combiner"
	if len(responseCombiners.data.Clone()) != 1 {
		t.Error("unexpected initial size of the response combiner list:", responseCombiners.data.Clone())
	}
	RegisterResponseCombiner(subject, getResponseCombiner(config.ExtraConfig{}))
	defer func() { responseCombiners = initResponseCombiners() }()

	if len(responseCombiners.data.Clone()) != 2 {
		t.Error("unexpected size of the response combiner list:", responseCombiners.data.Clone())
	}
	timeout := 500
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				mergeKey: defaultCombinerName,
			},
		},
	}
	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
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

func Test_incrementalMergeAccumulator_invalidResponse(t *testing.T) {
	acc := newIncrementalMergeAccumulator(3, combineData)
	acc.Merge(nil, nil)
	acc.Merge(nil, nil)
	acc.Merge(nil, nil)
	res, err := acc.Result()
	if res != nil {
		t.Error("response should be nil")
		return
	}
	if err == nil {
		t.Error("expecting error")
		return
	}
	switch mergeErr := err.(type) {
	case mergeError:
		if len(mergeErr.errs) != 3 {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
		if mergeErr.errs[0] != mergeErr.errs[1] {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
		if mergeErr.errs[0] != mergeErr.errs[2] {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
		if mergeErr.errs[0] != errNullResult {
			t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
		}
	default:
		t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
	}
}

func Test_incrementalMergeAccumulator_incompleteResponse(t *testing.T) {
	acc := newIncrementalMergeAccumulator(3, combineData)
	acc.Merge(&Response{Data: make(map[string]interface{}), IsComplete: true}, nil)
	acc.Merge(&Response{Data: make(map[string]interface{}), IsComplete: false}, nil)
	acc.Merge(&Response{Data: make(map[string]interface{}), IsComplete: true}, nil)
	res, err := acc.Result()
	if res == nil {
		t.Error("response should not be nil")
		return
	}
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
		return
	}
	if res.IsComplete {
		t.Error("response should not be completed")
	}
}

func testNewMergeDataMiddleware_simpleFiltering(t *testing.T) {
	timeout := 500
	backend := config.Backend{}
	endpoint := config.EndpointConfig{
		Backend: []*config.Backend{&backend, &backend},
		Timeout: time.Duration(timeout) * time.Millisecond,
	}

	RegisterBackendFiltererFactory("", func(_ *config.EndpointConfig) ([]BackendFilterer, error) {
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

func testNewMergeDataMiddleware_sequentialFiltering(t *testing.T) {
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

	RegisterBackendFiltererFactory("", func(_ *config.EndpointConfig) ([]BackendFilterer, error) {
		return []BackendFilterer{
			func(_ *Request) bool {
				return true
			},
			func(_ *Request) bool {
				return false
			},
			func(_ *Request) bool {
				return true
			},
		}, nil
	})

	mw := NewMergeDataMiddleware(logging.NoOp, &endpoint)
	p := mw(
		dummyProxy(&Response{Data: map[string]interface{}{"string": "some"}, IsComplete: true}),
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
		if len(out.Data) != 2 {
			t.Errorf("We were expecting a response from just two backends, but we got %v!\n", out)
		}
		if !out.IsComplete {
			t.Errorf("We were expecting a completed response but we got an incompleted one!\n")
		}
	}
}
