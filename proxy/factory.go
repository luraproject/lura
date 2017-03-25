package proxy

import (
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
)

// Factory creates proxies based on the received endpoint configuration.
//
// Both, factories and backend factories, create proxies but factories are designed as a stack makers
// because they are intended to generate the complete proxy stack for a given frontend endpoint
// the app would expose and they could wrap several proxies provided by a backend factory
type Factory interface {
	New(cfg *config.EndpointConfig) (Proxy, error)
}

// DefaultFactory returns a default http proxy factory with the injected logger
func DefaultFactory(logger logging.Logger) Factory {
	return NewDefaultFactory(httpProxy, logger)
}

// NewDefaultFactory returns a default proxy factory with the injected proxy builder and logger
func NewDefaultFactory(backendFactory BackendFactory, logger logging.Logger) Factory {
	return defaultFactory{backendFactory, logger}
}

type defaultFactory struct {
	backendFactory BackendFactory
	logger         logging.Logger
}

// New implements the Factory interface
func (pf defaultFactory) New(cfg *config.EndpointConfig) (p Proxy, err error) {
	switch len(cfg.Backend) {
	case 0:
		err = ErrNoBackends
	case 1:
		p, err = pf.newSingle(cfg)
	default:
		p, err = pf.newMulti(cfg)
	}
	return
}

func (pf defaultFactory) newMulti(cfg *config.EndpointConfig) (p Proxy, err error) {
	backendProxy := make([]Proxy, len(cfg.Backend))
	for i, backend := range cfg.Backend {
		backendProxy[i] = pf.newStack(backend)
	}
	p = NewMergeDataMiddleware(cfg)(backendProxy...)
	return
}

func (pf defaultFactory) newSingle(cfg *config.EndpointConfig) (Proxy, error) {
	return pf.newStack(cfg.Backend[0]), nil
}

func (pf defaultFactory) newStack(backend *config.Backend) (p Proxy) {
	p = pf.backendFactory(backend)
	p = NewRoundRobinLoadBalancedMiddleware(backend)(p)
	if backend.ConcurrentCalls > 1 {
		p = NewConcurrentMiddleware(backend)(p)
	}
	p = NewRequestBuilderMiddleware(backend)(p)
	return
}
