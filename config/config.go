// Package config defines the config structs and some config parser interfaces and implementations
package config

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/devopsfaith/krakend/encoding"
)

const (
	// BracketsRouterPatternBuilder uses brackets as route params delimiter
	BracketsRouterPatternBuilder = iota
	// ColonRouterPatternBuilder use a colon as route param delimiter
	ColonRouterPatternBuilder
	// DefaultMaxIdleConnsPerHost is the default value for the MaxIdleConnsPerHost param
	DefaultMaxIdleConnsPerHost = 250
	// DefaultTimeout is the default value to use for the ServiceConfig.Timeout param
	DefaultTimeout = 2 * time.Second

	// ConfigVersion is the current version of the config struct
	ConfigVersion = 2
)

// RoutingPattern to use during route conversion. By default, use the colon router pattern
var RoutingPattern = ColonRouterPatternBuilder

// ServiceConfig defines the krakend service
type ServiceConfig struct {
	// set of endpoint definitions
	Endpoints []*EndpointConfig `mapstructure:"endpoints"`
	// defafult timeout
	Timeout time.Duration `mapstructure:"timeout"`
	// default TTL for GET
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
	// default set of hosts
	Host []string `mapstructure:"host"`
	// port to bind the krakend service
	Port int `mapstructure:"port"`
	// version code of the configuration
	Version int `mapstructure:"version"`
	// Extra configuration for customized behaviour
	ExtraConfig ExtraConfig `mapstructure:"extra_config"`

	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`

	MaxIdleConnsPerHost int `mapstructure:"max_idle_connections"`

	// DisableStrictREST flags if the REST enforcement is disabled
	DisableStrictREST bool `mapstructure:"disable_rest"`

	// Plugin defines the configuration for the plugin loader
	Plugin *Plugin `mapstructure:"plugin"`

	// run krakend in debug mode
	Debug     bool
	uriParser URIParser
}

// EndpointConfig defines the configuration of a single endpoint to be exposed
// by the krakend service
type EndpointConfig struct {
	// url pattern to be registered and exposed to the world
	Endpoint string `mapstructure:"endpoint"`
	// HTTP method of the endpoint (GET, POST, PUT, etc)
	Method string `mapstructure:"method"`
	// set of definitions of the backends to be linked to this endpoint
	Backend []*Backend `mapstructure:"backend"`
	// number of concurrent calls this endpoint must send to the backends
	ConcurrentCalls int `mapstructure:"concurrent_calls"`
	// timeout of this endpoint
	Timeout time.Duration `mapstructure:"timeout"`
	// duration of the cache header
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
	// list of query string params to be extracted from the URI
	QueryString []string `mapstructure:"querystring_params"`
	// Endpoint Extra configuration for customized behaviour
	ExtraConfig ExtraConfig `mapstructure:"extra_config"`
	// HeadersToPass defines the list of headers to pass to the backends
	HeadersToPass []string `mapstructure:"headers_to_pass"`
}

// Backend defines how krakend should connect to the backend service (the API resource to consume)
// and how it should process the received response
type Backend struct {
	// the name of the group the response should be moved to. If empty, the response is
	// not changed
	Group string `mapstructure:"group"`
	// HTTP method of the request to send to the backend
	Method string `mapstructure:"method"`
	// Set of hosts of the API
	Host []string `mapstructure:"host"`
	// False if the hostname should be sanitized
	HostSanitizationDisabled bool `mapstructure:"disable_host_sanitize"`
	// URL pattern to use to locate the resource to be consumed
	URLPattern string `mapstructure:"url_pattern"`
	// set of response fields to remove. If empty, the filter id not used
	Blacklist []string `mapstructure:"blacklist"`
	// set of response fields to allow. If empty, the filter id not used
	Whitelist []string `mapstructure:"whitelist"`
	// map of response fields to be renamed and their new names
	Mapping map[string]string `mapstructure:"mapping"`
	// the encoding format
	Encoding string `mapstructure:"encoding"`
	// the response to process is a collection
	IsCollection bool `mapstructure:"is_collection"`
	// name of the field to extract to the root. If empty, the formater will do nothing
	Target string `mapstructure:"target"`
	// name of the service discovery driver to use
	SD string `mapstructure:"sd"`

	// list of keys to be replaced in the URLPattern
	URLKeys []string
	// number of concurrent calls this endpoint must send to the API
	ConcurrentCalls int
	// timeout of this backend
	Timeout time.Duration
	// decoder to use in order to parse the received response from the API
	Decoder encoding.Decoder
	// Backend Extra configuration for customized behaviours
	ExtraConfig ExtraConfig `mapstructure:"extra_config"`
}

// Plugin contains the config required by the plugin module
type Plugin struct {
	Folder  string `mapstructure:"folder"`
	Pattern string `mapstructure:"pattern"`
}

// ExtraConfig is a type to store extra configurations for customized behaviours
type ExtraConfig map[string]interface{}

// ConfigGetter is a function for parsing ExtraConfig into a previously know type
type ConfigGetter func(ExtraConfig) interface{}

// DefaultConfigGetter is the Default implementation for ConfigGetter, it just returns the ExtraConfig map.
func DefaultConfigGetter(extra ExtraConfig) interface{} { return extra }

const defaultNamespace = "github.com/devopsfaith/krakend/config"

// ConfigGetters map than match namespaces and ConfigGetter so the components knows which type to expect returned by the
// ConfigGetter ie: if we look for the defaultNamespace in the map, we will get the DefaultConfigGetter implementation
// which will return a ExtraConfig when called
var ConfigGetters = map[string]ConfigGetter{defaultNamespace: DefaultConfigGetter}

var (
	simpleURLKeysPattern = regexp.MustCompile(`\{([a-zA-Z\-_0-9]+)\}`)
	debugPattern         = "^[^/]|/__debug(/.*)?$"
	errInvalidHost       = errors.New("invalid host")
	defaultPort          = 8080
)

// Init initializes the configuration struct and its defined endpoints and backends.
// Init also sanitizes the values, applies the default ones whenever necessary and
// normalizes all the things.
func (s *ServiceConfig) Init() error {
	s.uriParser = NewURIParser()
	if s.Version != ConfigVersion {
		return fmt.Errorf("Unsupported version: %d (want: %d)", s.Version, ConfigVersion)
	}
	if s.Port == 0 {
		s.Port = defaultPort
	}
	if s.MaxIdleConnsPerHost == 0 {
		s.MaxIdleConnsPerHost = DefaultMaxIdleConnsPerHost
	}
	if s.Timeout == 0 {
		s.Timeout = DefaultTimeout
	}

	s.Host = s.uriParser.CleanHosts(s.Host)

	for i, e := range s.Endpoints {
		e.Endpoint = s.uriParser.CleanPath(e.Endpoint)

		if err := e.validate(); err != nil {
			return err
		}

		inputParams := s.extractPlaceHoldersFromURLTemplate(e.Endpoint, s.paramExtractionPattern())
		inputSet := map[string]interface{}{}
		for ip := range inputParams {
			inputSet[inputParams[ip]] = nil
		}

		e.Endpoint = s.uriParser.GetEndpointPath(e.Endpoint, inputParams)

		s.initEndpointDefaults(i)

		for j, b := range e.Backend {

			s.initBackendDefaults(i, j)

			b.Method = strings.ToTitle(b.Method)

			if err := s.initBackendURLMappings(i, j, inputSet); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *ServiceConfig) paramExtractionPattern() *regexp.Regexp {
	if s.DisableStrictREST {
		return simpleURLKeysPattern
	}
	return endpointURLKeysPattern
}

func (s *ServiceConfig) extractPlaceHoldersFromURLTemplate(subject string, pattern *regexp.Regexp) []string {
	matches := pattern.FindAllStringSubmatch(subject, -1)
	keys := make([]string, len(matches))
	for k, v := range matches {
		keys[k] = v[1]
	}
	return keys
}

func (s *ServiceConfig) initEndpointDefaults(e int) {
	endpoint := s.Endpoints[e]
	if endpoint.Method == "" {
		endpoint.Method = "GET"
	} else {
		endpoint.Method = strings.ToTitle(endpoint.Method)
	}
	if s.CacheTTL != 0 && endpoint.CacheTTL == 0 {
		endpoint.CacheTTL = s.CacheTTL
	}
	if s.Timeout != 0 && endpoint.Timeout == 0 {
		endpoint.Timeout = s.Timeout
	}
	if endpoint.ConcurrentCalls == 0 {
		endpoint.ConcurrentCalls = 1
	}
}

func (s *ServiceConfig) initBackendDefaults(e, b int) {
	endpoint := s.Endpoints[e]
	backend := endpoint.Backend[b]
	if len(backend.Host) == 0 {
		backend.Host = s.Host
	} else if !backend.HostSanitizationDisabled {
		backend.Host = s.uriParser.CleanHosts(backend.Host)
	}
	if backend.Method == "" {
		backend.Method = endpoint.Method
	}
	backend.Timeout = endpoint.Timeout
	backend.ConcurrentCalls = endpoint.ConcurrentCalls
	backend.Decoder = encoding.Get(strings.ToLower(backend.Encoding))(backend.IsCollection)
}

func (s *ServiceConfig) initBackendURLMappings(e, b int, inputParams map[string]interface{}) error {
	backend := s.Endpoints[e].Backend[b]

	backend.URLPattern = s.uriParser.CleanPath(backend.URLPattern)

	outputParams := s.extractPlaceHoldersFromURLTemplate(backend.URLPattern, simpleURLKeysPattern)

	outputSet := map[string]interface{}{}
	for op := range outputParams {
		outputSet[outputParams[op]] = nil
	}

	if len(outputSet) > len(inputParams) {
		return fmt.Errorf("Too many output params! input: %v, output: %v\n", outputSet, outputParams)
	}

	tmp := backend.URLPattern
	backend.URLKeys = make([]string, len(outputParams))
	for o := range outputParams {
		if _, ok := inputParams[outputParams[o]]; !ok {
			return fmt.Errorf("Undefined output param [%s]! input: %v, output: %v\n", outputParams[o], inputParams, outputParams)
		}
		tmp = strings.Replace(tmp, "{"+outputParams[o]+"}", "{{."+strings.Title(outputParams[o])+"}}", -1)
		backend.URLKeys = append(backend.URLKeys, strings.Title(outputParams[o]))
	}
	backend.URLPattern = tmp
	return nil
}

func (e *EndpointConfig) validate() error {
	matched, err := regexp.MatchString(debugPattern, e.Endpoint)
	if err != nil {
		log.Printf("ERROR: parsing the endpoint url [%s]: %s. Ignoring\n", e.Endpoint, err.Error())
		return err
	}
	if matched {
		return fmt.Errorf("ERROR: the endpoint url path [%s] is not a valid one!!! Ignoring\n", e.Endpoint)
	}

	if len(e.Backend) == 0 {
		return fmt.Errorf("WARNING: the [%s] endpoint has 0 backends defined! Ignoring\n", e.Endpoint)
	}
	return nil
}
