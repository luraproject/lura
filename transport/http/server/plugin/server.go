package plugin

import (
	"context"
	"net/http"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
)

const Namespace = "github_com/devopsfaith/krakend/transport/http/server/handler"

type RunServer func(context.Context, config.ServiceConfig, http.Handler) error

func New(logger logging.Logger, next RunServer) RunServer {
	return func(ctx context.Context, cfg config.ServiceConfig, handler http.Handler) error {
		v, ok := cfg.ExtraConfig[Namespace]
		if !ok {
			logger.Debug("http-server-handler: no extra config")
			return next(ctx, cfg, handler)
		}
		extra, ok := v.(map[string]interface{})
		if !ok {
			logger.Debug("http-server-handler: wrong extra config type")
			return next(ctx, cfg, handler)
		}

		// load plugin(s)
		r, ok := serverRegister.Get(Namespace)
		if !ok {
			logger.Debug("http-server-handler: no plugins registered for the module")
			return next(ctx, cfg, handler)
		}

		name, nameOk := extra["name"].(string)
		fifoRaw, fifoOk := extra["name"].([]interface{})
		if !nameOk && !fifoOk {
			logger.Debug("http-server-handler: no plugins required in the extra config")
			return next(ctx, cfg, handler)
		}
		fifo := []string{}

		if !fifoOk {
			fifo = []string{name}
		} else {
			for _, x := range fifoRaw {
				if v, ok := x.(string); ok {
					fifo = append(fifo, v)
				}
			}
		}

		for _, name := range fifo {
			rawHf, ok := r.Get(name)
			if !ok {
				logger.Debug("http-server-handler: no plugin registered as", name)
				return next(ctx, cfg, handler)
			}

			hf, hfOk := rawHf.(func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error))
			hfl, hflOk := rawHf.(func(context.Context, logging.Logger, map[string]interface{}, http.Handler) (http.Handler, error))

			var handlerWrapper http.Handler
			var err error

			if hfOk {
				handlerWrapper, err = hf(context.Background(), extra, handler)
			} else if hflOk {
				handlerWrapper, err = hfl(context.Background(), logger, extra, handler)
			} else {
				logger.Warning("http-server-handler: wrong plugin handler type:", name)
				return next(ctx, cfg, handler)
			}

			if err != nil {
				logger.Warning("http-server-handler: error getting the plugin handler:", err.Error())
				return next(ctx, cfg, handler)
			}

			logger.Debug("http-server-handler: injecting plugin", name)
			handler = handlerWrapper
		}
		return next(ctx, cfg, handler)
	}
}
