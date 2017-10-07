package martian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/google/martian/body"
	_ "github.com/google/martian/fifo"
	_ "github.com/google/martian/header"
	"github.com/google/martian/parse"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
)

func NewBackendFactory(logger logging.Logger, re proxy.HTTPRequestExecutor) proxy.BackendFactory {
	return func(remote *config.Backend) proxy.Proxy {
		martian, err := Parse(remote.ExtraConfig)
		switch err {
		case nil:
			return proxy.NewHTTPProxyWithHTTPExecutor(remote, HTTPRequestExecutor(martian, re), remote.Decoder)
		case ErrEmptyValue:
			return proxy.NewHTTPProxyWithHTTPExecutor(remote, re, remote.Decoder)
		default:
			logger.Error(err, remote.ExtraConfig)
			return proxy.NewHTTPProxyWithHTTPExecutor(remote, re, remote.Decoder)
		}
	}
}

func HTTPRequestExecutor(result *parse.Result, re proxy.HTTPRequestExecutor) proxy.HTTPRequestExecutor {
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		result.RequestModifier().ModifyRequest(req)
		resp, err := re(ctx, req)
		result.ResponseModifier().ModifyResponse(resp)
		return resp, err
	}
}

func Parse(e config.ExtraConfig) (*parse.Result, error) {
	cfg, ok := e[Namespace]
	if !ok {
		return nil, ErrEmptyValue
	}

	data, ok := cfg.(map[string]interface{})
	if !ok {
		return nil, ErrBadValue
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, ErrMarshallingValue
	}

	return parse.FromJSON(raw)
}

// Namespace is the key to look for extra configuration details
const Namespace = "github.com/krakend/proxy/martian"

var (
	ErrEmptyValue       = fmt.Errorf("getting the extra config for the martian proxy")
	ErrBadValue         = fmt.Errorf("casting the extra config for the martian proxy")
	ErrMarshallingValue = fmt.Errorf("marshalling the extra config for the martian proxy")
)
