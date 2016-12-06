package sd

import (
	"errors"
	"math/rand"
	"sync/atomic"
)

// Balancer applys a balancing stategy in order to select the backend host to be used
type Balancer interface {
	Host() (string, error)
}

// ErrNoHosts is the error the balancer must return when there are 0 hosts ready
var ErrNoHosts = errors.New("no hosts available")

// NewRoundRobinLB returns a new balancer using a round robin strategy
func NewRoundRobinLB(subscriber Subscriber) Balancer {
	return &roundRobinLB{
		subscriber: subscriber,
		counter:    0,
	}
}

type roundRobinLB struct {
	subscriber Subscriber
	counter    uint64
}

// Host implements the balancer interface
func (rr *roundRobinLB) Host() (string, error) {
	hosts, err := rr.subscriber.Hosts()
	if err != nil {
		return "", err
	}
	if len(hosts) <= 0 {
		return "", ErrNoHosts
	}
	offset := (atomic.AddUint64(&rr.counter, 1) - 1) % uint64(len(hosts))
	return hosts[offset], nil
}

// NewRandomLB returns a new balancer using a pseudo-random strategy
func NewRandomLB(subscriber Subscriber, seed int64) Balancer {
	return &randomLB{
		subscriber: subscriber,
		rnd:        rand.New(rand.NewSource(seed)),
	}
}

type randomLB struct {
	subscriber Subscriber
	rnd        *rand.Rand
}

// Host implements the balancer interface
func (r *randomLB) Host() (string, error) {
	hosts, err := r.subscriber.Hosts()
	if err != nil {
		return "", err
	}
	if len(hosts) <= 0 {
		return "", ErrNoHosts
	}
	return hosts[r.rnd.Intn(len(hosts))], nil
}
