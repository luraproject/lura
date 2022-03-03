// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestNewStaticMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()
	endpoint := config.EndpointConfig{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				staticKey: map[string]interface{}{
					"data": map[string]interface{}{
						"new-1": true,
						"new-2": map[string]interface{}{"k1": 42},
						"new-3": "42",
					},
					"strategy": "incomplete",
				},
			},
		},
	}
	mw := NewStaticMiddleware(logging.NoOp, &endpoint)
	mw(explosiveProxy(t), explosiveProxy(t))
}

func TestNewStaticMiddleware_ok(t *testing.T) {
	endpoint := config.EndpointConfig{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				staticKey: map[string]interface{}{
					"data": map[string]interface{}{
						"new-1": true,
						"new-2": map[string]interface{}{"k1": 42},
						"new-3": "42",
					},
					"strategy": "incomplete",
				},
			},
		},
	}
	mw := NewStaticMiddleware(logging.NoOp, &endpoint)

	p := mw(dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}, IsComplete: true}))
	out1, err := p(context.Background(), &Request{})
	if err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
	}
	if out1 == nil {
		t.Error("The proxy returned a null result")
		return
	}
	if len(out1.Data) != 1 {
		t.Errorf("We weren't expecting an extra partial response but we got %v!", out1)
	}
	if !out1.IsComplete {
		t.Errorf("We were expecting a completed response but we got an incompleted one!")
	}

	p = mw(dummyProxy(&Response{Data: map[string]interface{}{"supu": 42}}))
	out2, err := p(context.Background(), &Request{})
	if err != nil {
		t.Errorf("The middleware propagated an unexpected error: %s", err.Error())
	}
	if out2 == nil {
		t.Error("The proxy returned a null result")
		return
	}
	if len(out2.Data) != 4 {
		t.Errorf("We weren't expecting a partial response but we got %v!", out2)
	}

	expectedError := errors.New("expect me")
	p = mw(func(_ context.Context, _ *Request) (*Response, error) {
		return nil, expectedError
	})
	out3, err := p(context.Background(), &Request{})
	if err != expectedError {
		t.Errorf("The middleware propagated an unexpected error: %s", err)
	}
	if out3 == nil {
		t.Error("The proxy returned a null result")
		return
	}
	if len(out3.Data) != 3 {
		t.Errorf("We weren't expecting a partial response but we got %v!", out3)
	}
}

type staticMatcherTestCase struct {
	name     string
	response *Response
	err      error
	expected bool
}

func TestNewStaticMiddleware(t *testing.T) {
	data := map[string]interface{}{
		"new-1": true,
		"new-2": map[string]interface{}{"k1": 42},
		"new-3": "42",
	}
	extra := config.ExtraConfig{
		Namespace: map[string]interface{}{
			staticKey: map[string]interface{}{
				"data":     data,
				"strategy": staticIfCompleteStrategy,
			},
		},
	}

	mw := NewStaticMiddleware(logging.NoOp, &config.EndpointConfig{ExtraConfig: extra})

	p := mw(func(_ context.Context, r *Request) (*Response, error) {
		return &Response{IsComplete: true}, nil
	})

	resp, err := p(context.Background(), nil)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(data, resp.Data) {
		t.Errorf("unexpected data: %+v", resp.Data)
	}
}

func Test_staticAlwaysMatch(t *testing.T) {
	extra := config.ExtraConfig{
		Namespace: map[string]interface{}{
			staticKey: map[string]interface{}{
				"data": map[string]interface{}{
					"new-1": true,
					"new-2": map[string]interface{}{"k1": 42},
					"new-3": "42",
				},
			},
		},
	}
	cfg, _ := getStaticMiddlewareCfg(extra)
	for _, testCase := range []staticMatcherTestCase{
		{
			name:     "nil & nil",
			expected: true,
		},
		{
			name:     "nil & error",
			err:      errors.New("ignore me"),
			expected: true,
		},
		{
			name:     "complete & nil",
			response: &Response{Data: map[string]interface{}{}, IsComplete: true},
			expected: true,
		},
		{
			name:     "complete & error",
			response: &Response{Data: map[string]interface{}{}, IsComplete: true},
			err:      errors.New("ignore me"),
			expected: true,
		},
		{
			name:     "incomplete",
			response: &Response{},
			expected: true,
		},
	} {
		testStaticMatcher(t, cfg.Match, testCase)
	}
}

func Test_staticIfSuccessMatch(t *testing.T) {
	extra := config.ExtraConfig{
		Namespace: map[string]interface{}{
			staticKey: map[string]interface{}{
				"data": map[string]interface{}{
					"new-1": true,
					"new-2": map[string]interface{}{"k1": 42},
					"new-3": "42",
				},
				"strategy": staticIfSuccessStrategy,
			},
		},
	}
	cfg, _ := getStaticMiddlewareCfg(extra)
	for _, testCase := range []staticMatcherTestCase{
		{
			name:     "nil & nil",
			expected: true,
		},
		{
			name:     "nil & error",
			err:      errors.New("ignore me"),
			expected: false,
		},
		{
			name:     "success & nil",
			response: &Response{},
			expected: true,
		},
		{
			name:     "success & error",
			response: &Response{},
			err:      errors.New("ignore me"),
		},
	} {
		testStaticMatcher(t, cfg.Match, testCase)
	}
}

func Test_staticIfErroredMatch(t *testing.T) {
	extra := config.ExtraConfig{
		Namespace: map[string]interface{}{
			staticKey: map[string]interface{}{
				"data": map[string]interface{}{
					"new-1": true,
					"new-2": map[string]interface{}{"k1": 42},
					"new-3": "42",
				},
				"strategy": staticIfErroredStrategy,
			},
		},
	}
	cfg, _ := getStaticMiddlewareCfg(extra)
	for _, testCase := range []staticMatcherTestCase{
		{
			name: "nil & nil",
		},
		{
			name:     "nil & error",
			err:      errors.New("ignore me"),
			expected: true,
		},
		{
			name:     "success & nil",
			response: &Response{},
		},
		{
			name:     "success & error",
			response: &Response{},
			err:      errors.New("ignore me"),
			expected: true,
		},
	} {
		testStaticMatcher(t, cfg.Match, testCase)
	}
}

func Test_staticIfCompleteMatch(t *testing.T) {
	extra := config.ExtraConfig{
		Namespace: map[string]interface{}{
			staticKey: map[string]interface{}{
				"data": map[string]interface{}{
					"new-1": true,
					"new-2": map[string]interface{}{"k1": 42},
					"new-3": "42",
				},
				"strategy": staticIfCompleteStrategy,
			},
		},
	}
	cfg, _ := getStaticMiddlewareCfg(extra)
	for _, testCase := range []staticMatcherTestCase{
		{
			name: "nil & nil",
		},
		{
			name: "nil & error",
			err:  errors.New("ignore me"),
		},
		{
			name:     "success & nil",
			response: &Response{},
		},
		{
			name:     "success & error",
			response: &Response{},
			err:      errors.New("ignore me"),
		},
		{
			name:     "complete",
			response: &Response{IsComplete: true},
			expected: true,
		},
	} {
		testStaticMatcher(t, cfg.Match, testCase)
	}
}

func Test_staticIfIncompleteMatch(t *testing.T) {
	extra := config.ExtraConfig{
		Namespace: map[string]interface{}{
			staticKey: map[string]interface{}{
				"data": map[string]interface{}{
					"new-1": true,
					"new-2": map[string]interface{}{"k1": 42},
					"new-3": "42",
				},
				"strategy": staticIfIncompleteStrategy,
			},
		},
	}
	cfg, _ := getStaticMiddlewareCfg(extra)
	for _, testCase := range []staticMatcherTestCase{
		{
			name:     "nil & nil",
			expected: true,
		},
		{
			name:     "nil & error",
			err:      errors.New("ignore me"),
			expected: true,
		},
		{
			name:     "success & nil",
			response: &Response{},
			expected: true,
		},
		{
			name:     "success & error",
			response: &Response{},
			err:      errors.New("ignore me"),
			expected: true,
		},
		{
			name:     "complete",
			response: &Response{IsComplete: true},
		},
	} {
		testStaticMatcher(t, cfg.Match, testCase)
	}
}

func testStaticMatcher(t *testing.T, marcher func(*Response, error) bool, testCase staticMatcherTestCase) {
	if marcher(testCase.response, testCase.err) != testCase.expected {
		t.Errorf(
			"[%s] unexepecting match result (%v) with: %v, %v",
			testCase.name,
			testCase.expected,
			testCase.response,
			testCase.err,
		)
	}
}

func Test_getStaticMiddlewareCfg_ko(t *testing.T) {
	for i, cfg := range []config.ExtraConfig{
		{"a": 42},
		{Namespace: true},
		{Namespace: map[string]interface{}{}},
		{Namespace: map[string]interface{}{staticKey: 42}},
		{Namespace: map[string]interface{}{staticKey: map[string]interface{}{}}},
	} {
		if _, ok := getStaticMiddlewareCfg(cfg); ok {
			t.Errorf("expecting error on test #%d", i)
		}
	}
}

func Test_getStaticMiddlewareCfg_strategy(t *testing.T) {
	for _, strategy := range []string{
		staticAlwaysStrategy,
		staticIfSuccessStrategy,
		staticIfErroredStrategy,
		staticIfCompleteStrategy,
		staticIfIncompleteStrategy,
	} {
		cfg := config.ExtraConfig{
			Namespace: map[string]interface{}{
				staticKey: map[string]interface{}{
					"data":     map[string]interface{}{},
					"strategy": strategy,
				},
			},
		}
		staticCfg, ok := getStaticMiddlewareCfg(cfg)
		if !ok {
			t.Errorf("unexpecting error on test %s", strategy)
		}
		if strategy != staticCfg.Strategy {
			t.Errorf("wrong parsing on test %s", strategy)
		}
	}

	cfg := config.ExtraConfig{
		Namespace: map[string]interface{}{
			staticKey: map[string]interface{}{
				"data": map[string]interface{}{},
			},
		},
	}
	staticCfg, ok := getStaticMiddlewareCfg(cfg)
	if !ok {
		t.Error("unexpecting error parsing default strategy")
	}
	if staticAlwaysStrategy != staticCfg.Strategy {
		t.Error("wrong parsing default strategy")
	}
}
