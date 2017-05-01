package dnssrv

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
)

func TestSubscriber_New(t *testing.T) {
	srvSet := []*net.SRV{
		{
			Port:   80,
			Target: "127.0.0.1",
		},
		{
			Port:   81,
			Target: "127.0.0.1",
		},
	}
	DefaultLookup = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", srvSet, nil
	}

	s := SubscriberFactory(&config.Backend{Host: []string{"some.example.tld"}})
	hosts, err := s.Hosts()
	if err != nil {
		t.Error("Getting the hosts:", err.Error())
	}
	if len(hosts) != 2 {
		t.Error("Wrong number of hosts:", len(hosts))
	}
	if hosts[0] != "http://127.0.0.1:80" {
		t.Error("Wrong host #0 (expected http://127.0.0.1:80):", hosts[0])
	}
	if hosts[1] != "http://127.0.0.1:81" {
		t.Error("Wrong host #1 (expected http://127.0.0.1:81):", hosts[1])
	}
}

func TestSubscriber_LoockupError(t *testing.T) {
	errToReturn := errors.New("Some random error")
	defaultLookup := func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", []*net.SRV{}, errToReturn
	}
	ttl := 1 * time.Millisecond
	s := NewDetailed("some.example.tld", defaultLookup, ttl)
	hosts, err := s.Hosts()
	if err != nil {
		t.Error("Unexpected error!", err)
	}
	if len(hosts) != 0 {
		t.Error("Wrong number of hosts:", len(hosts))
	}
}
