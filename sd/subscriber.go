// Package sd defines some interfaces and implementations for service discovery
package sd

import (
	"github.com/devopsfaith/krakend/config"
)

// Subscriber keeps the set of backend hosts up to date
type Subscriber interface {
	Hosts() ([]string, error)
}

// FixedSubscriber has a constant set of backend hosts and they never get updated
type FixedSubscriber []string

// Hosts implements the subscriber interface
func (s FixedSubscriber) Hosts() ([]string, error) { return s, nil }

// SubscriberFactory builds subscribers with the received config
type SubscriberFactory func(*config.Backend) Subscriber

// FixedSubscriberFactory builds a FixedSubscriber with the received config
func FixedSubscriberFactory(cfg *config.Backend) Subscriber {
	return FixedSubscriber(cfg.Host)
}
