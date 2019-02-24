package http

import (
	"crypto/tls"
	"reflect"
	"testing"

	"github.com/devopsfaith/krakend/config"
)

func TestParseTLSConfig_nilConfig(t *testing.T) {
	if res := ParseTLSConfig(nil); res != nil {
		t.Errorf("unexpected result. have: %v, want nil", res)
	}
}

func TestParseTLSConfig_disabledConfig(t *testing.T) {
	if res := ParseTLSConfig(&config.TLS{IsDisabled: true}); res != nil {
		t.Errorf("unexpected result. have: %v, want nil", res)
	}
}

func TestParseTLSConfig_emptyConfig(t *testing.T) {
	res := ParseTLSConfig(&config.TLS{})
	if res == nil {
		t.Error("nil result")
		return
	}

	if !reflect.DeepEqual(res.CipherSuites, DefaultCipherSuites) {
		t.Errorf("unexpected cipher suites: %v", res.CipherSuites)
	}

	if !reflect.DeepEqual(res.CurvePreferences, DefaultCurves) {
		t.Errorf("unexpected cipher suites: %v", res.CurvePreferences)
	}

	if res.MinVersion != tls.VersionTLS12 || res.MaxVersion != tls.VersionTLS12 {
		t.Errorf("max and min version should be %v. min: %v, max: %v", tls.VersionTLS12, res.MinVersion, res.MaxVersion)
	}
}

func Test_parseTLSVersion(t *testing.T) {
	for _, tc := range []struct {
		in  string
		out uint16
	}{
		{in: "SSL3.0", out: tls.VersionSSL30},
		{in: "TLS10", out: tls.VersionTLS10},
		{in: "TLS11", out: tls.VersionTLS11},
		{in: "TLS12", out: tls.VersionTLS12},
		{in: "Unknown", out: tls.VersionTLS12},
	} {
		if res := parseTLSVersion(tc.in); res != tc.out {
			t.Errorf("input %s generated output %d. expected: %d", tc.in, res, tc.out)
		}
	}
}

func Test_parseCurveIDs(t *testing.T) {
	original := []uint16{1, 2, 3}
	cs := parseCurveIDs(&config.TLS{CurvePreferences: original})
	for k, v := range cs {
		if original[k] != uint16(v) {
			t.Errorf("unexpected curves %v. expected: %v", cs, original)
		}
	}
}

func Test_parseCipherSuites(t *testing.T) {
	original := []uint16{1, 2, 3}
	cs := parseCipherSuites(&config.TLS{CipherSuites: original})
	for k, v := range cs {
		if original[k] != uint16(v) {
			t.Errorf("unexpected ciphersuites %v. expected: %v", cs, original)
		}
	}
}
