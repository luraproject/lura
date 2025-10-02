package proxy

import (
    "github.com/luraproject/lura/config"
    "github.com/luraproject/lura/proxy/middleware"
)

// Factory builds proxies with middleware support
type Factory struct {
    baseFactory FactoryFunc
}

// FactoryFunc is the standard proxy factory function
type FactoryFunc func(cfg *config.Backend) Proxy

// NewFactory creates a factory with a given base function
func NewFactory(base FactoryFunc) *Factory {
    return &Factory{baseFactory: base}
}

// New builds a proxy with middlewares applied from config
func (f *Factory) New(cfg *config.Backend) Proxy {
    p := f.baseFactory(cfg)

    if raw, ok := cfg.ExtraConfig["middlewares"]; ok {
        if list, ok := raw.([]string); ok {
            for _, name := range list {
                if mw, exists := middleware.Get(name); exists {
                    p = mw(p) // wrap with middleware
                }
            }
        }
    }

    return p
}