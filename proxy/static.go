package proxy

import (
	"context"

	"github.com/devopsfaith/krakend/config"
)

// NewStaticMiddleware creates proxy middleware for adding static values to the processed responses
func NewStaticMiddleware(endpointConfig *config.EndpointConfig) Middleware {
	cfg, ok := getStaticMiddlewareCfg(endpointConfig.ExtraConfig)
	if !ok {
		return EmptyMiddleware
	}
	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, request *Request) (*Response, error) {
			result, err := next[0](ctx, request)
			if !cfg.Match(result, err) {
				return result, err
			}

			if result == nil {
				result = &Response{Data: map[string]interface{}{}}
			}

			for k, v := range cfg.Data {
				result.Data[k] = v
			}

			return result, err
		}
	}
}

const (
	staticKey = "static"

	staticAlwaysStrategy       = "always"
	staticIfSuccessStrategy    = "success"
	staticIfErroredStrategy    = "errored"
	staticIfCompleteStrategy   = "complete"
	staticIfIncompleteStrategy = "incomplete"
)

type staticConfig struct {
	Data     map[string]interface{}
	Strategy string
	Match    func(*Response, error) bool
}

func getStaticMiddlewareCfg(extra config.ExtraConfig) (staticConfig, bool) {
	v, ok := extra[Namespace]
	if !ok {
		return staticConfig{}, ok
	}
	e, ok := v.(map[string]interface{})
	if !ok {
		return staticConfig{}, ok
	}
	v, ok = e[staticKey]
	if !ok {
		return staticConfig{}, ok
	}
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return staticConfig{}, ok
	}
	data, ok := tmp["data"].(map[string]interface{})
	if !ok {
		return staticConfig{}, ok
	}

	name, ok := tmp["strategy"].(string)
	if !ok {
		name = staticAlwaysStrategy
	}
	cfg := staticConfig{
		Data:     data,
		Strategy: name,
		Match:    staticAlwaysMatch,
	}
	switch name {
	case staticAlwaysStrategy:
		cfg.Match = staticAlwaysMatch
	case staticIfSuccessStrategy:
		cfg.Match = staticIfSuccessMatch
	case staticIfErroredStrategy:
		cfg.Match = staticIfErroredMatch
	case staticIfCompleteStrategy:
		cfg.Match = staticIfCompleteMatch
	case staticIfIncompleteStrategy:
		cfg.Match = staticIfIncompleteMatch
	}
	return cfg, true
}

func staticAlwaysMatch(_ *Response, _ error) bool       { return true }
func staticIfSuccessMatch(_ *Response, err error) bool  { return err == nil }
func staticIfErroredMatch(_ *Response, err error) bool  { return err != nil }
func staticIfCompleteMatch(r *Response, err error) bool { return err == nil && r != nil && r.IsComplete }
func staticIfIncompleteMatch(r *Response, _ error) bool { return r == nil || !r.IsComplete }
