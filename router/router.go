// Package router defines some interfaces for router adapters
package router

import "github.com/devopsfaith/krakend/config"

// Router sets up the public layer exposed to the users
type Router interface {
	Run(cfg config.ServiceConfig)
}

// Factory creates new routers
type Factory interface {
	New() Router
}
