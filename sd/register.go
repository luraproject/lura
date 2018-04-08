package sd

import (
	"sync"

	"github.com/devopsfaith/krakend/config"
)

// Deprecated: RegisterSubscriberFactory. Use the GetRegister function
func RegisterSubscriberFactory(name string, sf SubscriberFactory) error {
	return subscriberFactories.Register(name, sf)
}

// Deprecated: GetSubscriber. Use the GetRegister function
func GetSubscriber(cfg *config.Backend) Subscriber {
	return subscriberFactories.Get(cfg.SD)(cfg)
}

// RegisterSetter registers the received subscriber factory for later usage
type RegisterSetter interface {
	Register(name string, sf SubscriberFactory) error
}

// RegisterGetter gets the subscriber factory by name or a fixed subscriber factory if
// the name is not registered
type RegisterGetter interface {
	Get(name string) SubscriberFactory
}

// GetRegister returns the package register
func GetRegister() *Register {
	return subscriberFactories
}

// Register is a SD register
type Register struct {
	data  map[string]SubscriberFactory
	mutex *sync.RWMutex
}

var subscriberFactories = &Register{
	data:  map[string]SubscriberFactory{},
	mutex: &sync.RWMutex{},
}

// Register implements the RegisterSetter interface
func (r *Register) Register(name string, sf SubscriberFactory) error {
	r.mutex.Lock()
	r.data[name] = sf
	r.mutex.Unlock()
	return nil
}

// Get implements the RegisterGetter interface
func (r *Register) Get(name string) SubscriberFactory {
	r.mutex.RLock()
	sf, ok := r.data[name]
	r.mutex.RUnlock()
	if !ok {
		return FixedSubscriberFactory
	}
	return sf
}
