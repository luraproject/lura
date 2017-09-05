package proxy

import (
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/sd"
	"github.com/devopsfaith/krakend/config"
)

var streamHttpProxy = StreamHTTPProxyFactory(NewHTTPClient)

// StreamDefaultFactory returns a default streaming http proxy factory with the injected logger, if the endpoint is not
// configured as streaming it will fallback to the DefaultFactory implementation
func StreamDefaultFactory(logger logging.Logger) Factory {
	return streamDefaultFactory{
		defaultFactory{streamHttpProxy, logger, sd.FixedSubscriberFactory},
		DefaultFactory(logger)}
}

type streamDefaultFactory struct {
	streamHttpFactory  defaultFactory
	defaultHttpFactory Factory
}

// New implements the Factory interface
func (pf streamDefaultFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	streamConfigGetter := config.ConfigGetters[config.StreamNamespace]
	streamExtraConfig := streamConfigGetter(cfg.ExtraConfig).(config.StreamExtraConfig)
	if streamExtraConfig.Forward {
		switch len(cfg.Backend) {
		case 0:
			err = ErrNoBackends
		case 1:
			return pf.streamHttpFactory.newSingle(cfg)
		default:
			err = ErrTooManyBackends
		}
		return
	} else {
		return pf.defaultHttpFactory.New(cfg)
	}

}
