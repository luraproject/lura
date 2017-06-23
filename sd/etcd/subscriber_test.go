package etcd

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/sd"
)

func TestSubscriberFactory_ko0Hosts(t *testing.T) {
	ctx := context.Background()
	c := dummyClient{
		getEntries:  func(key string) ([]string, error) { return []string{}, nil },
		watchPrefix: func(prefix string, ch chan struct{}) {},
	}

	var ops uint64

	fallbackSubscriberFactory = func(cfg *config.Backend) sd.Subscriber {
		atomic.AddUint64(&ops, 1)
		return sd.FixedSubscriberFactory(cfg)
	}

	SubscriberFactory(ctx, c)(&config.Backend{})

	if ops != 1 {
		t.Errorf("Unexpected number of calls to the fallback subscriber factory. Got: %d, Want: %d\n", ops, 1)
		return
	}
}

func TestSubscriberFactory_ko(t *testing.T) {
	ctx := context.Background()
	c := dummyClient{
		getEntries:  func(key string) ([]string, error) { return nil, fmt.Errorf("random fail") },
		watchPrefix: func(prefix string, ch chan struct{}) {},
	}

	var ops uint64

	fallbackSubscriberFactory = func(cfg *config.Backend) sd.Subscriber {
		atomic.AddUint64(&ops, 1)
		return sd.FixedSubscriberFactory(cfg)
	}

	conf := config.Backend{Host: []string{"random_etcd_service_name"}}
	SubscriberFactory(ctx, c)(&conf)

	if ops != 1 {
		t.Errorf("Unexpected number of calls to the fallback subscriber factory. Got: %d, Want: %d\n", ops, 1)
		return
	}
}

func TestSubscriberFactory_ok(t *testing.T) {
	ctx := context.Background()
	expectedHosts := []string{"first", "second", "third"}
	c := dummyClient{
		getEntries:  func(string) ([]string, error) { return expectedHosts, nil },
		watchPrefix: func(string, chan struct{}) {},
	}
	conf := config.Backend{Host: []string{"random_etcd_service_name"}}

	subscribers = map[string]sd.Subscriber{}

	sf := SubscriberFactory(ctx, c)
	if len(subscribers) != 0 {
		t.Errorf("Unexpected number of cached subscribers. Got: %d, Want: %d\n", len(subscribers), 0)
	}

	for i := 0; i < 4; i++ {
		hosts, err := sf(&conf).Hosts()
		if err != nil {
			t.Error(err)
			return
		}
		if len(subscribers) != 1 {
			t.Errorf("Unexpected number of cached subscribers. Got: %d, Want: %d\n", len(subscribers), 1)
		}

		if hosts[0] != expectedHosts[0] {
			t.Errorf("Unexpected number of calls to the fallback subscriber factory. Got: %v, Want: %v\n", hosts, expectedHosts)
			return
		}
	}
}

func TestNewSubscriber(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	expectedHosts := []string{"first", "second", "third"}
	lastSet := &[]string{}
	var fail bool
	shouldFail := &fail
	c := dummyClient{
		getEntries: func(key string) ([]string, error) {
			if *shouldFail {
				return nil, fmt.Errorf("random fail")
			}
			return *lastSet, nil
		},
		watchPrefix: func(prefix string, ch chan struct{}) {
			for {
				<-time.After(100 * time.Millisecond)
				*lastSet = expectedHosts
				*shouldFail = false
				ch <- struct{}{}
			}
		},
	}
	sb, err := NewSubscriber(ctx, c, "something")
	if err != nil {
		t.Error("Creating a subscriber:", err.Error())
		return
	}
	hs, err := sb.Hosts()
	if err != nil {
		t.Error("Getting hosts:", err.Error())
		return
	}
	if len(hs) != 0 {
		t.Errorf("Wrong initial number of hosts: %d\n", len(hs))
		return
	}
	<-time.After(100 * time.Millisecond)
	*shouldFail = true
	<-time.After(400 * time.Millisecond)
	hs, err = sb.Hosts()
	if err != nil {
		t.Error("Getting hosts:", err.Error())
		return
	}
	if len(hs) != len(expectedHosts) {
		t.Errorf("Wrong final number of hosts: %d\n", len(hs))
		return
	}
}

func TestNewSubscriber_ko(t *testing.T) {
	ctx := context.Background()
	c := dummyClient{
		getEntries:  func(key string) ([]string, error) { return nil, fmt.Errorf("random fail") },
		watchPrefix: func(prefix string, ch chan struct{}) {},
	}
	sb, err := NewSubscriber(ctx, c, "something")
	if err == nil {
		t.Error("Creating a subscriber:", err)
		return
	}
	if sb != nil {
		t.Error("Unexpected subscriber:", sb)
		return
	}
}

type dummyClient struct {
	getEntries  func(string) ([]string, error)
	watchPrefix func(string, chan struct{})
}

func (c dummyClient) GetEntries(key string) ([]string, error)     { return c.getEntries(key) }
func (c dummyClient) WatchPrefix(prefix string, ch chan struct{}) { c.watchPrefix(prefix, ch) }
