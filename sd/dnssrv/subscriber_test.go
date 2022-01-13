// SPDX-License-Identifier: Apache-2.0

package dnssrv

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/sd"
)

func ExampleRegister() {
	if err := Register(); err != nil {
		fmt.Println("registering the dns module:", err.Error())
		return
	}
	srvSet := []*net.SRV{
		{
			Port:   90,
			Target: "foobar",
			Weight: 2,
		},
		{
			Port:   90,
			Target: "127.0.0.1",
			Weight: 2,
		},
		{
			Port:   80,
			Target: "127.0.0.1",
			Weight: 2,
		},
		{
			Port:   81,
			Target: "127.0.0.1",
			Weight: 4,
		},
		{
			Port:     82,
			Target:   "127.0.0.1",
			Weight:   10,
			Priority: 2,
		},
		{
			Port:   83,
			Target: "127.0.0.1",
		},
	}
	DefaultLookup = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", srvSet, nil
	}

	s := sd.GetRegister().Get(Namespace)(&config.Backend{Host: []string{"some.example.tld"}, SD: Namespace})
	hosts, err := s.Hosts()
	if err != nil {
		fmt.Println("Getting the hosts:", err.Error())
		return
	}
	for _, h := range hosts {
		fmt.Println(h)
	}

	// output:
	// http://127.0.0.1:81
	// http://127.0.0.1:81
	// http://127.0.0.1:80
	// http://127.0.0.1:90
	// http://foobar:90
}

func ExampleNewDetailed() {
	srvSet := []*net.SRV{
		{
			Port:   90,
			Target: "foobar",
			Weight: 2,
		},
		{
			Port:   90,
			Target: "127.0.0.1",
			Weight: 2,
		},
		{
			Port:   80,
			Target: "127.0.0.1",
			Weight: 2,
		},
		{
			Port:   81,
			Target: "127.0.0.1",
			Weight: 4,
		},
		{
			Port:     82,
			Target:   "127.0.0.1",
			Weight:   10,
			Priority: 2,
		},
		{
			Port:   83,
			Target: "127.0.0.1",
		},
	}
	lookupFunc := func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", srvSet, nil
	}

	s := NewDetailed("some.example.tld", lookupFunc, 10*time.Second)
	hosts, err := s.Hosts()
	if err != nil {
		fmt.Println("Getting the hosts:", err.Error())
		return
	}
	for _, h := range hosts {
		fmt.Println(h)
	}

	// output:
	// http://127.0.0.1:81
	// http://127.0.0.1:81
	// http://127.0.0.1:80
	// http://127.0.0.1:90
	// http://foobar:90
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

func TestSubscriber_ResolveVeryLarge(t *testing.T) {
	var srvSet []*net.SRV
	const max = 1000
	for i := 0; i < max; i++ {
		srvSet = append(srvSet, &net.SRV{
			Port:   uint16(80 + i),
			Target: "127.0.0.1",
			Weight: 65535,
		})
	}
	lookupFunc := func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "cname", srvSet, nil
	}
	s := NewDetailed("large.example.tld", lookupFunc, 10*time.Second)
	hosts, _ := s.Hosts()
	if len(hosts) != max {
		t.Errorf("Expected %d, but got %d", max, len(hosts))
	}
}

func ExampleWeights_compact_basic() {
	for _, tc := range [][]uint16{
		[]uint16{25, 10000, 1000},
		[]uint16{25, 1000, 10000, 0, 65535},
		[]uint16{1, 65535},
		[]uint16{},
		[]uint16{0, 0, 0, 0},
	} {
		fmt.Println(tc, compact(tc))
	}

	// output:
	// [25 10000 1000] [0 10 1]
	// [25 1000 10000 0 65535] [0 1 13 0 85]
	// [1 65535] [0 1]
	// [] []
	// [0 0 0 0] [0 0 0 0]
}

func ExampleWeights_compact_custom() {
	tc := make([]uint16, 200)
	for i := range tc {
		tc[i] = uint16(3*5*7*11*13 + i)
	}
	fmt.Println(tc[:5], compact(tc[:5]))

	for i := range tc {
		tc[i] = uint16(i * 3 * 5 * 7)
	}
	fmt.Println(tc[:5], compact(tc[:5]))

	// output:
	// [15015 15016 15017 15018 15019] [19 19 20 20 20]
	// [0 105 210 315 420] [0 1 2 3 4]
}
