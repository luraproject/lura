package proxy

import (
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/sd"
	"github.com/devopsfaith/krakend/config"
)

const streamNamespace = "github.com/devopsfaith/krakend/config/stream"

func StreamConfigGetter(extra config.ExtraConfig) interface{} {
	ok := extra["Forward"];
	return StreamExtraConfig{ok != nil}
}

type StreamExtraConfig struct {
	forward bool
}

var streamHttpProxy = StreamHTTPProxyFactory(NewHTTPClient)

// DefaultFactory returns a default http proxy factory with the injected logger
func StreamDefaultFactory(logger logging.Logger) Factory {
	return streamDefaultFactory{defaultFactory{streamHttpProxy, logger, sd.FixedSubscriberFactory}}
}

type streamDefaultFactory struct {
	defaultFactory defaultFactory
}

// New implements the Factory interface
func (pf streamDefaultFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	streamConfigGetter := config.ConfigGetters[streamNamespace]
	streamExtraConfig := streamConfigGetter(cfg.ExtraConfig).(StreamExtraConfig)
	if streamExtraConfig.forward {
		switch len(cfg.Backend) {
		case 0:
			err = ErrNoBackends
		case 1:
			return pf.defaultFactory.newStack(cfg.Backend[0]), nil
		default:
			err = ErrTooManyBackends
		}
		return
	} else {
		return pf.defaultFactory.New(cfg)
	}

}
