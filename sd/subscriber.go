/* Package sd defines some interfaces and implementations for service discovery
 */
// SPDX-License-Identifier: Apache-2.0
package sd

import "github.com/luraproject/lura/config"

// Subscriber keeps the set of backend hosts up to date
type Subscriber interface {
	Hosts() ([]string, error)
}

// SubscriberFunc type is an adapter to allow the use of ordinary functions as subscribers.
// If f is a function with the appropriate signature, SubscriberFunc(f) is a Subscriber that calls f.
type SubscriberFunc func() ([]string, error)

// Hosts implements the Subscriber interface by executing the wrapped function
func (f SubscriberFunc) Hosts() ([]string, error) { return f() }

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
