// Package router defines some interfaces for router adapters
package router

import (
	"context"

	"github.com/devopsfaith/krakend/config"
)

// Router sets up the public layer exposed to the users
type Router interface {
	Run(config.ServiceConfig)
}

// FactoryFunc type is an adapter to allow the use of ordinary functions as routers.
// If f is a function with the appropriate signature, RouterFunc(f) is a Router that calls f.
type RouterFunc func(config.ServiceConfig)

// New implements the Router interface
func (f RouterFunc) Run(cfg config.ServiceConfig) { f(cfg) }

// Factory creates new routers
type Factory interface {
	New() Router
	NewWithContext(context.Context) Router
}
