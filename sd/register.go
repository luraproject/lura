package sd

import (
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/register"
)

// Deprecated: RegisterSubscriberFactory. Use the GetRegister function
func RegisterSubscriberFactory(name string, sf SubscriberFactory) error {
	return subscriberFactories.Register(name, sf)
}

// Deprecated: GetSubscriber. Use the GetRegister function
func GetSubscriber(cfg *config.Backend) Subscriber {
	return subscriberFactories.Get(cfg.SD)(cfg)
}

// GetRegister returns the package register
func GetRegister() *Register {
	return subscriberFactories
}

// Register is a SD register
type Register struct {
	data register.Untyped
}

func initRegister() *Register {
	return &Register{register.NewUntyped()}
}

// Register implements the RegisterSetter interface
func (r *Register) Register(name string, sf SubscriberFactory) error {
	r.data.Register(name, sf)
	return nil
}

// Get implements the RegisterGetter interface
func (r *Register) Get(name string) SubscriberFactory {
	tmp, ok := r.data.Get(name)
	if !ok {
		return FixedSubscriberFactory
	}
	sf, ok := tmp.(SubscriberFactory)
	if !ok {
		return FixedSubscriberFactory
	}
	return sf
}

var subscriberFactories = initRegister()
