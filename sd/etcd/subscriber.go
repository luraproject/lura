package etcd

import (
	"fmt"
	"sync"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/sd"
)

// SubscriberFactory builds a an etcd subscriber SubscriberFactory with the received etcd client
func SubscriberFactory(c Client) sd.SubscriberFactory {
	return func(cfg *config.Backend) sd.Subscriber {
		s, err := NewSubscriber(c, cfg.Host[0])
		if err != nil {
			return sd.FixedSubscriberFactory(cfg)
		}
		return s
	}
}

// Code taken from https://github.com/go-kit/kit/blob/master/sd/etcd/instancer.go

// Subscriber keeps instances stored in a certain etcd keyspace cached in a fixed subscriber. Any kind of
// change in that keyspace is watched and will update the Subscriber's list of hosts.
type Subscriber struct {
	cache  *sd.FixedSubscriber
	mutex  *sync.Mutex
	client Client
	prefix string
	quitc  chan struct{}
}

// NewSubscriber returns an etcd subscriber. It will start watching the given
// prefix for changes, and update the subscribers.
func NewSubscriber(c Client, prefix string) (*Subscriber, error) {
	s := &Subscriber{
		client: c,
		prefix: prefix,
		cache:  &sd.FixedSubscriber{},
		quitc:  make(chan struct{}),
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err == nil {
		fmt.Println("prefix", s.prefix, "instances", len(instances))
	} else {
		fmt.Println("prefix", s.prefix, "err", err)
	}
	*(s.cache) = sd.FixedSubscriber(instances)

	go s.loop()
	return s, nil
}

// Hosts implements the subscriber interface
func (s Subscriber) Hosts() ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.cache.Hosts()
}

func (s *Subscriber) loop() {
	ch := make(chan struct{})
	go s.client.WatchPrefix(s.prefix, ch)
	for {
		select {
		case <-ch:
			instances, err := s.client.GetEntries(s.prefix)
			if err != nil {
				fmt.Println("msg", "failed to retrieve entries", "err", err)
				continue
			}
			s.mutex.Lock()
			*(s.cache) = sd.FixedSubscriber(instances)
			s.mutex.Unlock()

		case <-s.quitc:
			return
		}
	}
}

// Stop terminates the Subscriber.
func (s *Subscriber) Stop() {
	close(s.quitc)
}
