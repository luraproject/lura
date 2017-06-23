package etcd

import (
	"context"
	"sync"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/sd"
)

var (
	subscribers               = map[string]sd.Subscriber{}
	subscribersMutex          = &sync.Mutex{}
	fallbackSubscriberFactory = sd.FixedSubscriberFactory
)

// SubscriberFactory builds a an etcd subscriber SubscriberFactory with the received etcd client
func SubscriberFactory(ctx context.Context, c Client) sd.SubscriberFactory {
	return func(cfg *config.Backend) sd.Subscriber {
		if len(cfg.Host) == 0 {
			return fallbackSubscriberFactory(cfg)
		}
		subscribersMutex.Lock()
		defer subscribersMutex.Unlock()
		if sf, ok := subscribers[cfg.Host[0]]; ok {
			return sf
		}
		sf, err := NewSubscriber(ctx, c, cfg.Host[0])
		if err != nil {
			return fallbackSubscriberFactory(cfg)
		}
		subscribers[cfg.Host[0]] = sf
		return sf
	}
}

// Code taken from https://github.com/go-kit/kit/blob/master/sd/etcd/instancer.go

// Subscriber keeps instances stored in a certain etcd keyspace cached in a fixed subscriber. Any kind of
// change in that keyspace is watched and will update the Subscriber's list of hosts.
type Subscriber struct {
	cache  *sd.FixedSubscriber
	mutex  *sync.RWMutex
	client Client
	prefix string
	ctx    context.Context
}

// NewSubscriber returns an etcd subscriber. It will start watching the given
// prefix for changes, and update the subscribers.
func NewSubscriber(ctx context.Context, c Client, prefix string) (*Subscriber, error) {
	s := &Subscriber{
		client: c,
		prefix: prefix,
		cache:  &sd.FixedSubscriber{},
		ctx:    ctx,
		mutex:  &sync.RWMutex{},
	}

	instances, err := s.client.GetEntries(s.prefix)
	if err != nil {
		return nil, err
	}
	*(s.cache) = sd.FixedSubscriber(instances)

	go s.loop()

	return s, nil
}

// Hosts implements the subscriber interface
func (s Subscriber) Hosts() ([]string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
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
				continue
			}
			s.mutex.Lock()
			*(s.cache) = sd.FixedSubscriber(instances)
			s.mutex.Unlock()

		case <-s.ctx.Done():
			return
		}
	}
}
