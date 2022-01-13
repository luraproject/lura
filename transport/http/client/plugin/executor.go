// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/transport/http/client"
)

const Namespace = "github.com/devopsfaith/krakend/transport/http/client/executor"

func HTTPRequestExecutor(
	logger logging.Logger,
	next func(*config.Backend) client.HTTPRequestExecutor,
) func(*config.Backend) client.HTTPRequestExecutor {
	return func(cfg *config.Backend) client.HTTPRequestExecutor {
		logPrefix := "[BACKEND: " + cfg.URLPattern + "]"
		v, ok := cfg.ExtraConfig[Namespace]
		if !ok {
			return next(cfg)
		}
		extra, ok := v.(map[string]interface{})
		if !ok {
			logger.Debug(logPrefix, "["+Namespace+"]", "Wrong extra config type for backend")
			return next(cfg)
		}

		// load plugin
		r, ok := clientRegister.Get(Namespace)
		if !ok {
			logger.Debug(logPrefix, "No plugins registered for the module")
			return next(cfg)
		}

		name, ok := extra["name"].(string)
		if !ok {
			logger.Debug(logPrefix, "No name defined in the extra config for", cfg.URLPattern)
			return next(cfg)
		}

		rawHf, ok := r.Get(name)
		if !ok {
			logger.Debug(logPrefix, "No plugin registered as", name)
			return next(cfg)
		}

		hf, ok := rawHf.(func(context.Context, map[string]interface{}) (http.Handler, error))
		if !ok {
			logger.Warning(logPrefix, "Wrong plugin handler type:", name)
			return next(cfg)
		}

		handler, err := hf(context.Background(), extra)
		if err != nil {
			logger.Warning(logPrefix, "Error getting the plugin handler:", err.Error())
			return next(cfg)
		}

		logger.Debug(logPrefix, "Injecting plugin", name)
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req.WithContext(ctx))
			return w.Result(), nil
		}
	}
}
