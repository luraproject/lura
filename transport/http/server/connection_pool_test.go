// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/http"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestNewDefaultTransport_LIFO(t *testing.T) {
	for _, strat := range []string{"", "lifo", "LIFO", " Lifo "} {
		rt := newDefaultTransport(config.ServiceConfig{ConnectionLeaseStrategy: strat}, logging.NoOp)
		if _, ok := rt.(*http.Transport); !ok {
			t.Errorf("strategy %q: expected plain *http.Transport, got %T", strat, rt)
		}
	}
}

func TestNewDefaultTransport_UnknownFallsBackToLIFO(t *testing.T) {
	rt := newDefaultTransport(config.ServiceConfig{ConnectionLeaseStrategy: "bogus"}, logging.NoOp)
	if _, ok := rt.(*http.Transport); !ok {
		t.Fatalf("expected plain *http.Transport for unknown strategy, got %T", rt)
	}
}

func TestNewDefaultTransport_FIFO(t *testing.T) {
	cfg := config.ServiceConfig{
		ConnectionLeaseStrategy: "fifo",
		ConnectionPools:         4,
		MaxIdleConns:            400,
		MaxIdleConnsPerHost:     100,
	}
	rt := newDefaultTransport(cfg, logging.NoOp)
	rr, ok := rt.(*roundRobinTransport)
	if !ok {
		t.Fatalf("expected *roundRobinTransport, got %T", rt)
	}
	if len(rr.transports) != 4 {
		t.Fatalf("ring size = %d, want 4", len(rr.transports))
	}
	for i, tr := range rr.transports {
		if tr.MaxIdleConns != 100 { // 400 / 4
			t.Errorf("pool %d MaxIdleConns = %d, want 100", i, tr.MaxIdleConns)
		}
		if tr.MaxIdleConnsPerHost != 25 { // 100 / 4
			t.Errorf("pool %d MaxIdleConnsPerHost = %d, want 25", i, tr.MaxIdleConnsPerHost)
		}
	}
}

func TestNewDefaultTransport_FIFODefaultsAndClamp(t *testing.T) {
	// pools <= 1 uses the default ring size.
	rt := newDefaultTransport(config.ServiceConfig{ConnectionLeaseStrategy: "fifo"}, logging.NoOp)
	if rr := rt.(*roundRobinTransport); len(rr.transports) != defaultConnectionPools {
		t.Errorf("default ring size = %d, want %d", len(rr.transports), defaultConnectionPools)
	}
	// oversized pools are clamped.
	rt = newDefaultTransport(config.ServiceConfig{ConnectionLeaseStrategy: "fifo", ConnectionPools: 10000}, logging.NoOp)
	if rr := rt.(*roundRobinTransport); len(rr.transports) != maxConnectionPools {
		t.Errorf("clamped ring size = %d, want %d", len(rr.transports), maxConnectionPools)
	}
}

func TestRoundRobinTransportRotates(t *testing.T) {
	const size = 3
	ring := make([]*http.Transport, size)
	for i := range ring {
		ring[i] = &http.Transport{}
	}
	rr := newRoundRobinTransport(ring)

	seen := make([]*http.Transport, 2*size)
	for i := range seen {
		seen[i] = rr.pick()
	}
	for i := 0; i < size; i++ {
		if seen[i] != seen[i+size] {
			t.Errorf("pick %d and %d differ; expected round-robin", i, i+size)
		}
	}
	if seen[0] == seen[1] || seen[1] == seen[2] || seen[0] == seen[2] {
		t.Errorf("expected %d distinct transports in the ring", size)
	}
}

func TestCeilDiv(t *testing.T) {
	cases := map[string][3]int{ // a, b -> want
		"exact":    {400, 4, 100},
		"round up": {250, 8, 32},
		"floor 1":  {3, 8, 1},
		"zero b":   {250, 0, 250},
	}
	for name, c := range cases {
		if got := ceilDiv(c[0], c[1]); got != c[2] {
			t.Errorf("%s: ceilDiv(%d,%d) = %d, want %d", name, c[0], c[1], got, c[2])
		}
	}
}
