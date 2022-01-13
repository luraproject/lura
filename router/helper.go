// SPDX-License-Identifier: Apache-2.0

package router

import (
	"net/http"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
)

func IsValidSequentialEndpoint(endpoint *config.EndpointConfig) bool {
	if endpoint.ExtraConfig[proxy.Namespace] == nil {
		return false
	}

	proxyCfg := endpoint.ExtraConfig[proxy.Namespace].(map[string]interface{})
	if proxyCfg["sequential"] == false {
		return false
	}

	for i, backend := range endpoint.Backend {
		if backend.Method != http.MethodGet && (i+1) != len(endpoint.Backend) {
			return false
		}
	}

	return true
}
