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
		return result, fmt.Errorf("Fatal error config file: %s \n", configFile)
	}
	if err = json.Unmarshal(data, &cfg); err != nil {
		return result, fmt.Errorf("Fatal error config file: While parsing config: %s \n", err.Error())
	}
	result = cfg.normalize()
	err = result.Init()

	return result, err
}

type parseableServiceConfig struct {
	Endpoints   []*parseableEndpointConfig `mapstructure:"endpoints"`
	Timeout     string                     `mapstructure:"timeout"`
	CacheTTL    string                     `mapstructure:"cache_ttl"`
	Host        []string                   `mapstructure:"host"`
	Port        int                        `mapstructure:"port"`
	Version     int                        `mapstructure:"version"`
	ExtraConfig *ExtraConfig               `mapstructure:"extra_config,omitempty" json:"extra_config,omitempty"`
	Debug       bool
}

func (p *parseableServiceConfig) normalize() ServiceConfig {
	cfg := ServiceConfig{
		Timeout:  parseDuration(p.Timeout),
		CacheTTL: parseDuration(p.CacheTTL),
		Host:     p.Host,
		Port:     p.Port,
		Version:  p.Version,
		Debug:    p.Debug,
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
	Endpoint        string              `mapstructure:"endpoint"`
	Method          string              `mapstructure:"method"`
	Backend         []*parseableBackend `mapstructure:"backend"`
	ConcurrentCalls int                 `mapstructure:"concurrent_calls"`
	Timeout         string              `mapstructure:"timeout"`
	CacheTTL        string              `mapstructure:"cache_ttl"`
	QueryString     []string            `mapstructure:"querystring_params"`
	ExtraConfig     *ExtraConfig        `mapstructure:"extra_config,omitempty" json:"extra_config,omitempty"`
}

func (p *parseableEndpointConfig) normalize() *EndpointConfig {
	e := EndpointConfig{
		Endpoint:        p.Endpoint,
		Method:          p.Method,
		ConcurrentCalls: p.ConcurrentCalls,
		Timeout:         parseDuration(p.Timeout),
		CacheTTL:        parseDuration(p.CacheTTL),
		QueryString:     p.QueryString,
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
	Group                    string            `mapstructure:"group"`
	Method                   string            `mapstructure:"method"`
	Host                     []string          `mapstructure:"host"`
	HostSanitizationDisabled bool              `mapstructure:"disable_host_sanitize"`
	URLPattern               string            `mapstructure:"url_pattern"`
	Blacklist                []string          `mapstructure:"blacklist"`
	Whitelist                []string          `mapstructure:"whitelist"`
	Mapping                  map[string]string `mapstructure:"mapping"`
	Encoding                 string            `mapstructure:"encoding"`
	IsCollection             bool              `mapstructure:"is_collection"`
	Target                   string            `mapstructure:"target"`
	ExtraConfig              *ExtraConfig      `mapstructure:"extra_config,omitempty" json:"extra_config,omitempty"`
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
