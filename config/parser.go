package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// Parser reads a configuration file, parses it and returns the content as an init ServiceConfig struct
type Parser interface {
	Parse(configFile string) (ServiceConfig, error)
}

// ParserFunc type is an adapter to allow the use of ordinary functions as subscribers.
// If f is a function with the appropriate signature, ParserFunc(f) is a Parser that calls f.
type ParserFunc func(string) (ServiceConfig, error)

// Parse implements the Parser interface
func (f ParserFunc) Parse(configFile string) (ServiceConfig, error) { return f(configFile) }

// NewParser creates a new parser using the json library
func NewParser() Parser {
	return parser{}
}

type parser struct{}

// Parser implements the Parse interface
func (p parser) Parse(configFile string) (ServiceConfig, error) {
	var result ServiceConfig
	var cfg parseableServiceConfig
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return result, fmt.Errorf("Fatal error config file: %s", configFile)
	}
	if err = json.Unmarshal(data, &cfg); err != nil {
		return result, fmt.Errorf("Fatal error config file: While parsing config: %s", err.Error())
	}
	result = cfg.normalize()
	err = result.Init()

	return result, err
}

type parseableServiceConfig struct {
	Endpoints             []*parseableEndpointConfig `json:"endpoints"`
	Timeout               string                     `json:"timeout"`
	CacheTTL              string                     `json:"cache_ttl"`
	Host                  []string                   `json:"host"`
	Port                  int                        `json:"port"`
	Version               int                        `json:"version"`
	ExtraConfig           *ExtraConfig               `json:"extra_config,omitempty"`
	ReadTimeout           string                     `json:"read_timeout"`
	WriteTimeout          string                     `json:"write_timeout"`
	IdleTimeout           string                     `json:"idle_timeout"`
	ReadHeaderTimeout     string                     `json:"read_header_timeout"`
	DisableKeepAlives     bool                       `json:"disable_keep_alives"`
	DisableCompression    bool                       `json:"disable_compression"`
	MaxIdleConns          int                        `json:"max_idle_connections"`
	MaxIdleConnsPerHost   int                        `json:"max_idle_connections_per_host"`
	IdleConnTimeout       string                     `json:"idle_connection_timeout"`
	ResponseHeaderTimeout string                     `json:"response_header_timeout"`
	ExpectContinueTimeout string                     `json:"expect_continue_timeout"`
	OutputEncoding        string                     `json:"output_encoding"`
	DialerTimeout         string                     `json:"dialer_timeout"`
	DialerFallbackDelay   string                     `json:"dialer_fallback_delay"`
	DialerKeepAlive       string                     `json:"dialer_keep_alive"`
	Debug                 bool
	Plugin                *Plugin
}

func (p *parseableServiceConfig) normalize() ServiceConfig {
	cfg := ServiceConfig{
		Timeout:               parseDuration(p.Timeout),
		CacheTTL:              parseDuration(p.CacheTTL),
		Host:                  p.Host,
		Port:                  p.Port,
		Version:               p.Version,
		Debug:                 p.Debug,
		ReadTimeout:           parseDuration(p.ReadTimeout),
		WriteTimeout:          parseDuration(p.WriteTimeout),
		IdleTimeout:           parseDuration(p.IdleTimeout),
		ReadHeaderTimeout:     parseDuration(p.ReadHeaderTimeout),
		DisableKeepAlives:     p.DisableKeepAlives,
		DisableCompression:    p.DisableCompression,
		MaxIdleConns:          p.MaxIdleConns,
		MaxIdleConnsPerHost:   p.MaxIdleConnsPerHost,
		IdleConnTimeout:       parseDuration(p.IdleConnTimeout),
		ResponseHeaderTimeout: parseDuration(p.ResponseHeaderTimeout),
		ExpectContinueTimeout: parseDuration(p.ExpectContinueTimeout),
		DialerTimeout:         parseDuration(p.DialerTimeout),
		DialerFallbackDelay:   parseDuration(p.DialerFallbackDelay),
		DialerKeepAlive:       parseDuration(p.DialerKeepAlive),
		OutputEncoding:        p.OutputEncoding,
		Plugin:                p.Plugin,
	}
	if p.ExtraConfig != nil {
		cfg.ExtraConfig = *p.ExtraConfig
	}
	endpoints := []*EndpointConfig{}
	for _, e := range p.Endpoints {
		endpoints = append(endpoints, e.normalize())
	}
	cfg.Endpoints = endpoints
	return cfg
}

type parseableEndpointConfig struct {
	Endpoint        string              `json:"endpoint"`
	Method          string              `json:"method"`
	Backend         []*parseableBackend `json:"backend"`
	ConcurrentCalls int                 `json:"concurrent_calls"`
	Timeout         string              `json:"timeout"`
	CacheTTL        int                 `json:"cache_ttl"`
	QueryString     []string            `json:"querystring_params"`
	ExtraConfig     *ExtraConfig        `json:"extra_config,omitempty"`
	HeadersToPass   []string            `json:"headers_to_pass"`
	OutputEncoding  string              `json:"output_encoding"`
}

func (p *parseableEndpointConfig) normalize() *EndpointConfig {
	e := EndpointConfig{
		Endpoint:        p.Endpoint,
		Method:          p.Method,
		ConcurrentCalls: p.ConcurrentCalls,
		Timeout:         parseDuration(p.Timeout),
		CacheTTL:        time.Duration(p.CacheTTL) * time.Second,
		QueryString:     p.QueryString,
		HeadersToPass:   p.HeadersToPass,
		OutputEncoding:  p.OutputEncoding,
	}
	if p.ExtraConfig != nil {
		e.ExtraConfig = *p.ExtraConfig
	}
	backends := []*Backend{}
	for _, b := range p.Backend {
		backends = append(backends, b.normalize())
	}
	e.Backend = backends
	return &e
}

type parseableBackend struct {
	Group                    string            `json:"group"`
	Method                   string            `json:"method"`
	Host                     []string          `json:"host"`
	HostSanitizationDisabled bool              `json:"disable_host_sanitize"`
	URLPattern               string            `json:"url_pattern"`
	Blacklist                []string          `json:"blacklist"`
	Whitelist                []string          `json:"whitelist"`
	Mapping                  map[string]string `json:"mapping"`
	Encoding                 string            `json:"encoding"`
	IsCollection             bool              `json:"is_collection"`
	Target                   string            `json:"target"`
	ExtraConfig              *ExtraConfig      `json:"extra_config,omitempty"`
	SD                       string            `json:"sd"`
}

func (p *parseableBackend) normalize() *Backend {
	b := Backend{
		Group:  p.Group,
		Method: p.Method,
		Host:   p.Host,
		HostSanitizationDisabled: p.HostSanitizationDisabled,
		URLPattern:               p.URLPattern,
		Blacklist:                p.Blacklist,
		Whitelist:                p.Whitelist,
		Mapping:                  p.Mapping,
		Encoding:                 p.Encoding,
		IsCollection:             p.IsCollection,
		Target:                   p.Target,
		SD:                       p.SD,
	}
	if p.ExtraConfig != nil {
		b.ExtraConfig = *p.ExtraConfig
	}
	return &b
}

func parseDuration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0
	}
	return d
}
