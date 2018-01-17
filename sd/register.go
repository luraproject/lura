package sd

import "github.com/devopsfaith/krakend/config"

// RegisterSubscriberFactory registers the received subscriber factory for later usage
func RegisterSubscriberFactory(name string, sf SubscriberFactory) error {
	subscriberFactories[name] = sf
	return nil
}

// GetSubscriber gets the subscriber factory by name or a fixed subscriber factory if
// the name is not registered
func GetSubscriber(cfg *config.Backend) Subscriber {
	sf, ok := subscriberFactories[cfg.SD]
	if !ok {
		return FixedSubscriberFactory(cfg)
	}
	return sf(cfg)
}

var subscriberFactories = map[string]SubscriberFactory{}
