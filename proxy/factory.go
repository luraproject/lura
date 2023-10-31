// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"fmt"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/sd"
)

// Factory creates proxies based on the received endpoint configuration.
//
// Both, factories and backend factories, create proxies but factories are designed as a stack makers
// because they are intended to generate the complete proxy stack for a given frontend endpoint
// the app would expose and they could wrap several proxies provided by a backend factory
type Factory interface {
	New(cfg *config.EndpointConfig) (Proxy, error)
}

// FactoryFunc type is an adapter to allow the use of ordinary functions as proxy factories.
// If f is a function with the appropriate signature, FactoryFunc(f) is a Factory that calls f.
type FactoryFunc func(*config.EndpointConfig) (Proxy, error)

// New implements the Factory interface
func (f FactoryFunc) New(cfg *config.EndpointConfig) (Proxy, error) { return f(cfg) }

// DefaultFactory returns a default http proxy factory with the injected logger
func DefaultFactory(logger logging.Logger) Factory {
	return NewDefaultFactory(httpProxy, logger)
}

// DefaultFactoryWithSubscriber returns a default proxy factory with the injected logger and subscriber factory
func DefaultFactoryWithSubscriber(logger logging.Logger, sF sd.SubscriberFactory) Factory {
	return NewDefaultFactoryWithSubscriber(httpProxy, logger, sF)
}

// NewDefaultFactory returns a default proxy factory with the injected proxy builder and logger
func NewDefaultFactory(backendFactory BackendFactory, logger logging.Logger) Factory {
	sf := func(remote *config.Backend) sd.Subscriber {
		return sd.GetRegister().Get(remote.SD)(remote)
	}
	return NewDefaultFactoryWithSubscriber(backendFactory, logger, sf)
}

// NewDefaultFactoryWithSubscriber returns a default proxy factory with the injected proxy builder,
// logger and subscriber factory
func NewDefaultFactoryWithSubscriber(backendFactory BackendFactory, logger logging.Logger, sF sd.SubscriberFactory) Factory {
	return defaultFactory{backendFactory, logger, sF}
}

type defaultFactory struct {
	backendFactory    BackendFactory
	logger            logging.Logger
	subscriberFactory sd.SubscriberFactory
}

// New implements the Factory interface
func (pf defaultFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	l, _ := logging.NewAnnotatedLogger(pf.logger, fmt.Sprintf("%s %s", cfg.Method, cfg.Endpoint))
	switch len(cfg.Backend) {
	case 0:
		err = ErrNoBackends
	case 1:
		p, err = pf.newSingle(l, cfg)
	default:
		p, err = pf.newMulti(l, cfg)
	}
	if err != nil {
		return
	}

	p = NewPluginMiddleware(l, cfg)(p)
	p = NewStaticMiddleware(l, cfg)(p)
	return
}

func (pf defaultFactory) newMulti(l logging.Logger, cfg *config.EndpointConfig) (p Proxy, err error) {
	backendProxy := make([]Proxy, len(cfg.Backend))
	for i, backend := range cfg.Backend {
		backendProxy[i] = pf.newStack(l, backend)
	}
	p = NewMergeDataMiddleware(l, cfg)(backendProxy...)
	p = NewFlatmapMiddleware(l, cfg)(p)
	return
}

func (pf defaultFactory) newSingle(l logging.Logger, cfg *config.EndpointConfig) (Proxy, error) {
	return pf.newStack(l, cfg.Backend[0]), nil
}

func (pf defaultFactory) newStack(l logging.Logger, backend *config.Backend) (p Proxy) {
	p = pf.backendFactory(backend)
	p = NewBackendPluginMiddleware(l, backend)(p)
	p = NewGraphQLMiddleware(l, backend)(p)
	p = NewFilterHeadersMiddleware(l, backend)(p)
	p = NewFilterQueryStringsMiddleware(l, backend)(p)
	p = NewLoadBalancedMiddlewareWithSubscriberAndLogger(l, pf.subscriberFactory(backend))(p)
	if backend.ConcurrentCalls > 1 {
		p = NewConcurrentMiddlewareWithLogger(l, backend)(p)
	}
	p = NewRequestBuilderMiddlewareWithLogger(l, backend)(p)
	return
}
