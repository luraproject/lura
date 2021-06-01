/* Package dnssrv defines some implementations for a dns based service discovery
 */
// SPDX-License-Identifier: Apache-2.0
package dnssrv

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/sd"
)

// Namespace is the key for the dns sd module
const Namespace = "dns"

// Register registers the dns sd subscriber factory
func Register() error {
	return sd.RegisterSubscriberFactory(Namespace, SubscriberFactory)
}

var (
	// TTL is the duration of the cached data
	TTL = 30 * time.Second
	// DefaultLookup id the function for the DNS resolution
	DefaultLookup = net.LookupSRV
)

// SubscriberFactory builds a DNS_SRV Subscriber with the received config
func SubscriberFactory(cfg *config.Backend) sd.Subscriber {
	return New(cfg.Host[0])
}

// New creates a DNS subscriber with the default values
func New(name string) sd.Subscriber {
	return NewDetailed(name, DefaultLookup, TTL)
}

// NewDetailed creates a DNS subscriber with the received values
func NewDetailed(name string, lookup lookup, ttl time.Duration) sd.Subscriber {
	s := subscriber{name, &sd.FixedSubscriber{}, &sync.Mutex{}, ttl, lookup}
	s.update()
	go s.loop()
	return s
}

type lookup func(service, proto, name string) (cname string, addrs []*net.SRV, err error)

type subscriber struct {
	name   string
	cache  *sd.FixedSubscriber
	mutex  *sync.Mutex
	ttl    time.Duration
	lookup lookup
}

// Hosts implements the subscriber interface
func (s subscriber) Hosts() ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.cache.Hosts()
}

func (s subscriber) loop() {
	for {
		<-time.After(s.ttl)
		s.update()
	}
}

func (s subscriber) update() {
	instances, err := s.resolve()
	if err != nil {
		return
	}
	s.mutex.Lock()
	*(s.cache) = sd.FixedSubscriber(instances)
	s.mutex.Unlock()
}

func (s subscriber) resolve() ([]string, error) {
	_, addrs, err := s.lookup("", "", s.name)
	if err != nil {
		return []string{}, err
	}
	instances := []string{}
	for _, addr := range addrs {
		instances = append(instances, fmt.Sprintf("http://%s", net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))))
		for i := 0; i < int(addr.Weight-1); i++ {
			instances = append(instances, fmt.Sprintf("http://%s", net.JoinHostPort(addr.Target, fmt.Sprint(addr.Port))))
		}
	}
	return instances, nil
}
