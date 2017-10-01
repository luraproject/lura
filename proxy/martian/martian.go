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
		if err != nil {
			logger.Error(err)
			return proxy.NewHTTPProxyWithHTTPExecutor(remote, re, remote.Decoder)
		}
		return proxy.NewHTTPProxyWithHTTPExecutor(remote, HTTPRequestExecutor(martian, re), remote.Decoder)
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
		return nil, fmt.Errorf("getting the extra config for the martian proxy")
	}

	data, ok := cfg.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("casting the extra config for the martian proxy")
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshalling the extra config for the martian proxy")
	}

	return parse.FromJSON(raw)
}

// Namespace is the key to look for extra configuration details
const Namespace = "github.com/krakend/proxy/martian"
