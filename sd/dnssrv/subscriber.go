// SPDX-License-Identifier: Apache-2.0

/*
	Package dnssrv defines some implementations for a dns based service discovery
*/
package dnssrv

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/sd"
)

// Namespace is the key for the dns sd module
const Namespace = "dns"

// Register registers the dns sd subscriber factory under the name defined by Namespace
func Register() error {
	return sd.GetRegister().Register(Namespace, SubscriberFactory)
}

// TTL is the duration of the cached data
var TTL = 30 * time.Second

// DefaultLookup is the function used for the DNS resolution
var DefaultLookup = net.LookupSRV

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
	s := subscriber{
		name:   name,
		cache:  &sd.FixedSubscriber{},
		mutex:  &sync.RWMutex{},
		ttl:    ttl,
		lookup: lookup,
	}

	s.update()

	go func() {
		for {
			<-time.After(s.ttl)
			s.update()
		}
	}()

	return s
}

type lookup func(service, proto, name string) (cname string, addrs []*net.SRV, err error)

type subscriber struct {
	name   string
	cache  *sd.FixedSubscriber
	mutex  *sync.RWMutex
	ttl    time.Duration
	lookup lookup
}

// Hosts returns a copy of the cached set of hosts. It is safe to call it concurrently
func (s subscriber) Hosts() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	hs, err := s.cache.Hosts()
	if err != nil {
		return []string{}, err
	}

	res := make([]string, len(hs))
	copy(res, hs)
	return res, nil
}

func (s subscriber) update() {
	instances, err := s.resolve()
	if err != nil {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(instances) > 100 {
		*(s.cache) = sd.NewRandomFixedSubscriber(instances)
	} else {
		*(s.cache) = sd.FixedSubscriber(instances)
	}
}

func (s subscriber) resolve() ([]string, error) {
	_, srvs, err := s.lookup("", "", s.name)
	if err != nil {
		return []string{}, err
	}

	sort.Slice(
		srvs,
		func(i, j int) bool {
			if srvs[i].Priority == srvs[j].Priority {
				if srvs[i].Weight == srvs[j].Weight {
					if srvs[i].Target == srvs[j].Target {
						return srvs[i].Port < srvs[j].Port
					}
					return srvs[i].Target < srvs[j].Target
				}
				return srvs[i].Weight > srvs[j].Weight
			}
			return srvs[i].Priority < srvs[j].Priority
		},
	)

	ws := []uint16{}
	host := []string{}

	for _, a := range srvs {
		if a.Priority > srvs[0].Priority {
			break
		}
		ws = append(ws, a.Weight)
		host = append(host, "http://"+net.JoinHostPort(a.Target, fmt.Sprint(a.Port)))
	}

	instances := []string{}
	for i, times := range compact(ws) {
		for j := uint16(0); j < times; j++ {
			instances = append(instances, host[i])
		}
	}
	return instances, nil
}

func compact(ws []uint16) []uint16 {
	tmp := normalize(ws)
	div := gcd(tmp)
	if div < 2 {
		return tmp
	}

	res := make([]uint16, len(tmp))
	for i, w := range tmp {
		res[i] = w / div
	}

	return res
}

func normalize(ws []uint16) []uint16 {
	scale := 100
	if l := len(ws); l > scale {
		scale = l
	}

	var sum int64
	for _, w := range ws {
		sum += int64(w)
	}
	if sum <= int64(scale) {
		return ws
	}

	res := make([]uint16, len(ws))
	for i, w := range ws {
		res[i] = uint16(int64(w) * int64(scale) / sum)
	}
	return res
}

func gcd(ws []uint16) uint16 {
	if len(ws) == 0 {
		return 0
	}

	localGCD := func(a uint16, b uint16) uint16 {
		for b > 0 {
			a, b = b, a%b
		}
		return a
	}

	result := ws[0]
	for _, i := range ws[1:] {
		result = localGCD(result, i)
	}

	return result
}
