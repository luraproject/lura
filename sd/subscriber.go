// Package sd defines some interfaces and implementations for service discovery
package sd

// Subscriber keeps the set of backend hosts up to date
type Subscriber interface {
	Hosts() ([]string, error)
}

// FixedSubscriber has a constant set of backend hosts and they never get updated
type FixedSubscriber []string

// Hosts implements the subscriber interface
func (s FixedSubscriber) Hosts() ([]string, error) { return s, nil }
