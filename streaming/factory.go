package streaming

import (
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/sd"
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
)

// StreamDefaultFactory returns a default streaming http proxy factory with the injected logger, if the endpoint is not
// configured as streaming it will fallback to the DefaultFactory implementation
func StreamDefaultFactory(logger logging.Logger) proxy.Factory {
	return streamDefaultFactory{
		proxy.NewDefaultFactoryWithSubscriber(streamHttpProxy, logger, sd.FixedSubscriberFactory),
		proxy.DefaultFactory(logger)}
}

type streamDefaultFactory struct {
	streamHttpFactory  proxy.Factory
	defaultHttpFactory proxy.Factory
}

// New implements the Factory interface
func (pf streamDefaultFactory) New(cfg *config.EndpointConfig) (p proxy.Proxy, err error) {
	streamConfigGetter := config.ConfigGetters[StreamNamespace]
	streamExtraConfig := streamConfigGetter(cfg.ExtraConfig).(StreamExtraConfig)
	if streamExtraConfig.Forward {
		switch len(cfg.Backend) {
		case 0:
			err = proxy.ErrNoBackends
		case 1:
			return pf.streamHttpFactory.New(cfg)
		default:
			err = proxy.ErrTooManyBackends
		}
		return
	} else {
		return pf.defaultHttpFactory.New(cfg)
	}

}
