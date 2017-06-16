package etcd

import (
	"context"
	"testing"
	"time"
)

func TestNewSubscriber(t *testing.T) {
	ctx := context.Background()
	expectedHosts := []string{"first", "second", "third"}
	lastSet := &[]string{}
	c := dummyClient{
		getEntries: func(key string) ([]string, error) { return *lastSet, nil },
		watchPrefix: func(prefix string, ch chan struct{}) {
			for {
				<-time.After(100 * time.Millisecond)
				*lastSet = expectedHosts
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
		t.Errorf("Wrong initial number of hosts: %d", len(hs))
		return
	}
	<-time.After(500 * time.Millisecond)
	hs, err = sb.Hosts()
	if err != nil {
		t.Error("Getting hosts:", err.Error())
		return
	}
	if len(hs) != len(expectedHosts) {
		t.Errorf("Wrong final number of hosts: %d", len(hs))
		return
	}
}

type dummyClient struct {
	getEntries  func(string) ([]string, error)
	watchPrefix func(string, chan struct{})
}

func (c dummyClient) GetEntries(key string) ([]string, error)     { return c.getEntries(key) }
func (c dummyClient) WatchPrefix(prefix string, ch chan struct{}) { c.watchPrefix(prefix, ch) }
