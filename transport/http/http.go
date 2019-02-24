package http

import (
	"crypto/tls"

	"github.com/devopsfaith/krakend/config"
)

// ParseTLSConfig creates a tls.Config from the TLS section of the service configuration
func ParseTLSConfig(cfg *config.TLS) *tls.Config {
	if cfg == nil {
		return nil
	}
	if cfg.IsDisabled {
		return nil
	}

	return &tls.Config{
		MinVersion:               parseTLSVersion(cfg.MinVersion),
		MaxVersion:               parseTLSVersion(cfg.MaxVersion),
		CurvePreferences:         parseCurveIDs(cfg),
		PreferServerCipherSuites: cfg.PreferServerCipherSuites,
		CipherSuites:             parseCipherSuites(cfg),
	}
}

func parseTLSVersion(key string) uint16 {
	if v, ok := TLSVersions[key]; ok {
		return v
	}
	return tls.VersionTLS12
}

func parseCurveIDs(cfg *config.TLS) []tls.CurveID {
	l := len(cfg.CurvePreferences)
	if l == 0 {
		return DefaultCurves
	}

	curves := make([]tls.CurveID, len(cfg.CurvePreferences))
	for i := range curves {
		curves[i] = tls.CurveID(cfg.CurvePreferences[i])
	}
	return curves
}

func parseCipherSuites(cfg *config.TLS) []uint16 {
	l := len(cfg.CipherSuites)
	if l == 0 {
		return DefaultCipherSuites
	}

	cs := make([]uint16, l)
	for i := range cs {
		cs[i] = uint16(cfg.CipherSuites[i])
	}
	return cs
}

var (
	// DefaultCurves are the preferred curve ids
	DefaultCurves = []tls.CurveID{
		tls.CurveP521,
		tls.CurveP384,
		tls.CurveP256,
	}
	// DefaultCipherSuites contains the preferred ciphersuites
	DefaultCipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}
	// TLSVersions is a map between tls versions and their ids
	TLSVersions = map[string]uint16{
		"SSL3.0": tls.VersionSSL30,
		"TLS10":  tls.VersionTLS10,
		"TLS11":  tls.VersionTLS11,
		"TLS12":  tls.VersionTLS12,
	}
)
