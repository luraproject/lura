// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

const (
	// leaseStrategyLIFO is the Go net/http default: the most-recently-used idle
	// connection is reused first (stack). It is the behaviour when the strategy
	// is unset, so the plain transport is used for it.
	leaseStrategyLIFO = "lifo"
	// leaseStrategyFIFO round-robins requests across a ring of independent
	// transports so idle connections are cycled evenly and traffic is spread
	// across backend nodes (queue-like behaviour). Go's stdlib transport cannot
	// do FIFO within a single idle pool, so we round-robin over N pools to obtain
	// the even-distribution benefits.
	leaseStrategyFIFO = "fifo"

	// defaultConnectionPools is the ring size used by the "fifo" strategy when
	// ConnectionPools is not set. Eight suits a handful of backend replicas while
	// keeping strong keep-alive reuse under the default max_idle_connections_per_host.
	defaultConnectionPools = 8
	// maxConnectionPools is an upper safety clamp so a mistyped config cannot
	// spawn an unreasonable number of transports/sockets.
	maxConnectionPools = 256
)

// newDefaultTransport builds the RoundTripper installed as http.DefaultTransport
// from the service config. For the default "lifo" strategy it returns the plain
// stdlib transport; for "fifo" it returns a round-robin ring of independent
// transports cloned from it.
func newDefaultTransport(cfg config.ServiceConfig, logger logging.Logger) http.RoundTripper {
	base := NewTransport(cfg, logger)

	switch strings.ToLower(strings.TrimSpace(cfg.ConnectionLeaseStrategy)) {
	case "", leaseStrategyLIFO:
		return base
	case leaseStrategyFIFO:
		// handled below
	default:
		logger.Warning(fmt.Sprintf("%s unknown connection_lease_strategy %q, falling back to lifo",
			loggerPrefix, cfg.ConnectionLeaseStrategy))
		return base
	}

	size := cfg.ConnectionPools
	if size <= 1 {
		size = defaultConnectionPools
	}
	if size > maxConnectionPools {
		logger.Warning(fmt.Sprintf("%s connection_pools %d exceeds the maximum %d, clamping",
			loggerPrefix, size, maxConnectionPools))
		size = maxConnectionPools
	}
	// A ring wider than the per-host idle cap starves each pool of keep-alive
	// reuse (every pool holds a single idle connection and re-dials per request).
	if base.MaxIdleConnsPerHost > 0 && size > base.MaxIdleConnsPerHost {
		logger.Warning(fmt.Sprintf("%s connection_pools %d exceeds max_idle_connections_per_host %d; each pool keeps a single idle connection, reducing keep-alive reuse",
			loggerPrefix, size, base.MaxIdleConnsPerHost))
	}

	ring := make([]*http.Transport, size)
	for i := range ring {
		t := base.Clone()
		// Keep the aggregate idle-pool capacity roughly aligned with the
		// configured limits by splitting them across the ring. A value of 0
		// means "unlimited" in net/http, so it is left untouched.
		if base.MaxIdleConns > 0 {
			t.MaxIdleConns = ceilDiv(base.MaxIdleConns, size)
		}
		if base.MaxIdleConnsPerHost > 0 {
			t.MaxIdleConnsPerHost = ceilDiv(base.MaxIdleConnsPerHost, size)
		}
		ring[i] = t
	}

	logger.Debug(fmt.Sprintf("%s FIFO connection leasing enabled (ring of %d transports)", loggerPrefix, size))
	return newRoundRobinTransport(ring)
}

// ceilDiv returns ceil(a/b) with a floor of 1, guarding against a zero divisor.
func ceilDiv(a, b int) int {
	if b <= 0 {
		return a
	}
	r := a / b
	if a%b != 0 {
		r++
	}
	if r < 1 {
		r = 1
	}
	return r
}

// roundRobinTransport dispatches each request to the next transport in the ring
// using an atomic counter. Each transport keeps its own idle-connection pool, so
// spreading requests across the ring cycles connections evenly and distributes
// traffic across backend nodes instead of always reusing the warmest connection.
type roundRobinTransport struct {
	transports []*http.Transport
	counter    uint64
}

func newRoundRobinTransport(transports []*http.Transport) *roundRobinTransport {
	return &roundRobinTransport{transports: transports}
}

// pick returns the next transport in the ring.
func (rr *roundRobinTransport) pick() *http.Transport {
	n := atomic.AddUint64(&rr.counter, 1) - 1
	return rr.transports[int(n%uint64(len(rr.transports)))]
}

// RoundTrip implements http.RoundTripper.
func (rr *roundRobinTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return rr.pick().RoundTrip(req)
}

// CloseIdleConnections closes idle connections on every transport in the ring so
// the ring honours the http.Client.CloseIdleConnections contract.
func (rr *roundRobinTransport) CloseIdleConnections() {
	for _, t := range rr.transports {
		t.CloseIdleConnections()
	}
}
