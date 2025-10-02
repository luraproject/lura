package middleware

import "github.com/luraproject/lura/proxy"

// ProxyMiddleware wraps a proxy function
type ProxyMiddleware func(proxy.Proxy) proxy.Proxy

var registry = map[string]ProxyMiddleware{}

// Register allows custom middlewares to be added
func Register(name string, mw ProxyMiddleware) {
    registry[name] = mw
}

// Get retrieves a middleware by name
func Get(name string) (ProxyMiddleware, bool) {
    mw, ok := registry[name]
    return mw, ok
}