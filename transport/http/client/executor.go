package client

import (
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/devopsfaith/krakend/config"
	krakendhttp "github.com/devopsfaith/krakend/transport/http"
)

// HTTPRequestExecutor defines the interface of the request executor for the HTTP transport protocol
type HTTPRequestExecutor func(ctx context.Context, req *http.Request) (*http.Response, error)

// DefaultHTTPRequestExecutor creates a HTTPRequestExecutor with the received HTTPClientFactory
func DefaultHTTPRequestExecutor(clientFactory HTTPClientFactory) HTTPRequestExecutor {
	return func(ctx context.Context, req *http.Request) (*http.Response, error) {
		return clientFactory(ctx).Do(req.WithContext(ctx))
	}
}

// HTTPClientFactory creates http clients based with the received context
type HTTPClientFactory func(ctx context.Context) *http.Client

// NewHTTPClient just returns the http default client
func NewHTTPClient(ctx context.Context) *http.Client { return defaultHTTPClient }

var defaultHTTPClient = &http.Client{}

var once sync.Once

// InitHTTPDefaultTransport inits the transport component of the default HTTP client
func InitHTTPDefaultTransport(cfg config.ServiceConfig) error {
	var err error
	once.Do(func() {
		var dc *http.Client
		dc, err = newHTTPClient(cfg)
		if dc != nil {
			defaultHTTPClient = dc
		}
	})
	return err
}

func newHTTPClient(cfg config.ServiceConfig) (*http.Client, error) {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
	}

	if cfg.TLS == nil || cfg.TLS.IsDisabled {
		return &http.Client{
			Transport: transport,
		}, nil
	}

	transport.TLSClientConfig = krakendhttp.ParseTLSConfig(cfg.TLS)

	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if cfg.TLS.LocalCA != "" {
		certs, err := ioutil.ReadFile(cfg.TLS.LocalCA)
		if err != nil {
			return nil, fmt.Errorf("Failed to append %s to RootCAs: %v", cfg.TLS.LocalCA, err)
		}
		rootCAs.AppendCertsFromPEM(certs)
	}

	transport.TLSClientConfig.RootCAs = rootCAs

	return &http.Client{
		Transport: transport,
	}, nil
}
