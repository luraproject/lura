// SPDX-License-Identifier: Apache-2.0

package sd

import (
	"github.com/luraproject/lura/v2/register"
)

// GetRegister returns the package register
func GetRegister() *Register {
	return subscriberFactories
}

type untypedRegister interface {
	Register(name string, v interface{})
	Get(name string) (interface{}, bool)
}

// Register is a SD register, mapping different SD subscriber factories
// to their respective name, so they can be accessed by name
type Register struct {
	data untypedRegister
}

func initRegister() *Register {
	return &Register{register.NewUntyped()}
}

// Register adds the SubscriberFactory to the internal register under the given
// name
func (r *Register) Register(name string, sf SubscriberFactory) error {
	r.data.Register(name, sf)
	return nil
}

// Get returns the SubscriberFactory stored under the given name. It falls back to
// a FixedSubscriberFactory if there is no factory with that name
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
