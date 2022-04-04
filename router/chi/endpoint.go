// SPDX-License-Identifier: Apache-2.0

package chi

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/router/mux"
)

// HandlerFactory creates a handler function that adapts the chi router with the injected proxy
type HandlerFactory func(*config.EndpointConfig, proxy.Proxy) http.HandlerFunc

// NewEndpointHandler implements the HandleFactory interface using the default ToHTTPError function
func NewEndpointHandler(cfg *config.EndpointConfig, prxy proxy.Proxy) http.HandlerFunc {
	hf := mux.CustomEndpointHandler(
		mux.NewRequestBuilder(extractParamsFromEndpoint),
	)
	return hf(cfg, prxy)
}

func extractParamsFromEndpoint(r *http.Request) map[string]string {
	ctx := r.Context()
	rctx := chi.RouteContext(ctx)

	params := map[string]string{}
	if len(rctx.URLParams.Keys) > 0 {
		for _, param := range rctx.URLParams.Keys {
			params[strings.Title(param[:1])+param[1:]] = chi.URLParam(r, param)
		}
	}
	return params
}
