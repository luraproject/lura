package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/core"
	krakendhttp "github.com/devopsfaith/krakend/transport/http"
)

// ToHTTPError translates an error into a HTTP status code
type ToHTTPError func(error) int

// DefaultToHTTPError is a ToHTTPError transalator that always returns an
// internal server error
func DefaultToHTTPError(_ error) int {
	return http.StatusInternalServerError
}

const (
	// HeaderCompleteResponseValue is the value of the CompleteResponseHeader
	// if the response is complete
	HeaderCompleteResponseValue = "true"
	// HeaderIncompleteResponseValue is the value of the CompleteResponseHeader
	// if the response is not complete
	HeaderIncompleteResponseValue = "false"
)

var (
	// CompleteResponseHeaderName is the header to flag incomplete responses to the client
	CompleteResponseHeaderName = "X-KrakenD-Completed"
	// HeadersToSend are the headers to pass from the router request to the proxy
	HeadersToSend = []string{"Content-Type"}
	// UserAgentHeaderValue is the value of the User-Agent header to add to the proxy request
	UserAgentHeaderValue = []string{core.KrakendUserAgent}

	// ErrInternalError is the error returned by the router when something went wrong
	ErrInternalError = errors.New("internal server error")
	// ErrPrivateKey is the error returned by the router when the private key is not defined
	ErrPrivateKey = errors.New("private key not defined")
	// ErrPublicKey is the error returned by the router when the public key is not defined
	ErrPublicKey = errors.New("public key not defined")
)

// InitHTTPDefaultTransport ensures the default HTTP transport is configured just once per execution
func InitHTTPDefaultTransport(cfg config.ServiceConfig) {
	onceTransportConfig.Do(func() {
		http.DefaultTransport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:       cfg.DialerTimeout,
				KeepAlive:     cfg.DialerKeepAlive,
				FallbackDelay: cfg.DialerFallbackDelay,
				DualStack:     true,
			}).DialContext,
			DisableCompression:    cfg.DisableCompression,
			DisableKeepAlives:     cfg.DisableKeepAlives,
			MaxIdleConns:          cfg.MaxIdleConns,
			MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
			IdleConnTimeout:       cfg.IdleConnTimeout,
			ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
			ExpectContinueTimeout: cfg.ExpectContinueTimeout,
			TLSHandshakeTimeout:   10 * time.Second,
		}
	})
}

// RunServer runs a http.Server with the given handler and configuration.
// It configures the TLS layer if required by the received configuration.
func RunServer(ctx context.Context, cfg config.ServiceConfig, handler http.Handler) error {
	done := make(chan error)
	s := NewServer(cfg, handler)

	if s.TLSConfig == nil {
		go func() {
			done <- s.ListenAndServe()
		}()
	} else {
		if cfg.TLS.PublicKey == "" {
			return ErrPublicKey
		}
		if cfg.TLS.PrivateKey == "" {
			return ErrPrivateKey
		}
		go func() {
			done <- s.ListenAndServeTLS(cfg.TLS.PublicKey, cfg.TLS.PrivateKey)
		}()
	}

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	}
}

// NewServer returns a http.Server ready to serve the injected handler
func NewServer(cfg config.ServiceConfig, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		TLSConfig:         krakendhttp.ParseTLSConfig(cfg.TLS),
	}
}

// ParseTLSConfig creates a tls.Config from the TLS section of the service configuration
func ParseTLSConfig(cfg *config.TLS) *tls.Config {
	return krakendhttp.ParseTLSConfig(cfg)
}

var onceTransportConfig sync.Once
