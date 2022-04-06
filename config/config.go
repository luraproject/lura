// SPDX-License-Identifier: Apache-2.0

/*
	Package config defines the config structs and some config parser interfaces and implementations
*/
package config

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/luraproject/lura/v2/encoding"
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
	ConfigVersion = 3
)

// RoutingPattern to use during route conversion. By default, use the colon router pattern
var RoutingPattern = ColonRouterPatternBuilder

// ServiceConfig defines the lura service
type ServiceConfig struct {
	// name of the service
	Name string `mapstructure:"name"`
	// set of endpoint definitions
	Endpoints []*EndpointConfig `mapstructure:"endpoints"`
	// set of async agent definitions
	AsyncAgents []*AsyncAgent `mapstructure:"async_agent"`
	// defafult timeout
	Timeout time.Duration `mapstructure:"timeout"`
	// default TTL for GET
	CacheTTL time.Duration `mapstructure:"cache_ttl"`
	// default set of hosts
	Host []string `mapstructure:"host"`
	// port to bind the lura service
	Port int `mapstructure:"port"`
	// version code of the configuration
	Version int `mapstructure:"version"`
	// OutputEncoding defines the default encoding strategy to use for the endpoint responses
	OutputEncoding string `mapstructure:"output_encoding"`
	// Extra configuration for customized behaviour
	ExtraConfig ExtraConfig `mapstructure:"extra_config"`

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration `mapstructure:"read_timeout"`
	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If IdleTimeout
	// is zero, the value of ReadTimeout is used. If both are
	// zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body.
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`

	// DisableKeepAlives, if true, prevents re-use of TCP connections
	// between different HTTP requests.
	DisableKeepAlives bool `mapstructure:"disable_keep_alives"`
	// DisableCompression, if true, prevents the Transport from
	// requesting compression with an "Accept-Encoding: gzip"
	// request header when the Request contains no existing
	// Accept-Encoding value. If the Transport requests gzip on
	// its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user
	// explicitly requested gzip it is not automatically
	// uncompressed.
	DisableCompression bool `mapstructure:"disable_compression"`
	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns int `mapstructure:"max_idle_connections"`
	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `mapstructure:"max_idle_connections_per_host"`
	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	IdleConnTimeout time.Duration `mapstructure:"idle_connection_timeout"`
	// ResponseHeaderTimeout, if non-zero, specifies the amount of
	// time to wait for a server's response headers after fully
	// writing the request (including its body, if any). This
	// time does not include the time to read the response body.
	ResponseHeaderTimeout time.Duration `mapstructure:"response_header_timeout"`
	// ExpectContinueTimeout, if non-zero, specifies the amount of
	// time to wait for a server's first response headers after fully
	// writing the request headers if the request has an
	// "Expect: 100-continue" header. Zero means no timeout and
	// causes the body to be sent immediately, without
	// waiting for the server to approve.
	// This time does not include the time to send the request header.
	ExpectContinueTimeout time.Duration `mapstructure:"expect_continue_timeout"`
	// DialerTimeout is the maximum amount of time a dial will wait for
	// a connect to complete. If Deadline is also set, it may fail
	// earlier.
	//
	// The default is no timeout.
	//
	// When using TCP and dialing a host name with multiple IP
	// addresses, the timeout may be divided between them.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	DialerTimeout time.Duration `mapstructure:"dialer_timeout"`
	// DialerFallbackDelay specifies the length of time to wait before
	// spawning a fallback connection, when DualStack is enabled.
	// If zero, a default delay of 300ms is used.
	DialerFallbackDelay time.Duration `mapstructure:"dialer_fallback_delay"`
	// DialerKeepAlive specifies the keep-alive period for an active
	// network connection.
	// If zero, keep-alives are not enabled. Network protocols
	// that do not support keep-alives ignore this field.
	DialerKeepAlive time.Duration `mapstructure:"dialer_keep_alive"`

	// DisableStrictREST flags if the REST enforcement is disabled
	DisableStrictREST bool `mapstructure:"disable_rest"`

	// Plugin defines the configuration for the plugin loader
	Plugin *Plugin `mapstructure:"plugin"`

	// TLS defines the configuration params for enabling TLS (HTTPS & HTTP/2) at
	// the router layer
	TLS *TLS `mapstructure:"tls"`

	// run lura in debug mode
	Debug     bool
	uriParser URIParser

	// SequentialStart flags if the agents should be started sequentially
	// before starting the router
	SequentialStart bool `mapstructure:"sequential_start"`
}

// AsyncAgent defines the configuration of a single subscriber/consumer to be initialized
// and maintained by the lura service
type AsyncAgent struct {
	Name       string     `mapstructure:"name"`
	Connection Connection `mapstructure:"connection"`
	Consumer   Consumer   `mapstructure:"consumer"`
	// the encoding format
	Encoding string `mapstructure:"encoding"`
	// set of definitions of the backends to be linked to this endpoint
	Backend []*Backend `mapstructure:"backend"`

	// Endpoint Extra configuration for customized behaviour
	ExtraConfig ExtraConfig `mapstructure:"extra_config"`
}

type Consumer struct {
	// timeout of the pipe defined by this subscriber
	Timeout time.Duration `mapstructure:"timeout"`
	Workers int           `mapstructure:"workers"`
	Topic   string        `mapstructure:"topic"`
	MaxRate float64       `mapstructure:"max_rate"`
}

type Connection struct {
	MaxRetries      int           `mapstructure:"max_retries"`
	BackoffStrategy string        `mapstructure:"backoff_strategy"`
	HealthInterval  time.Duration `mapstructure:"health_interval"`
}

// EndpointConfig defines the configuration of a single endpoint to be exposed
// by the lura service
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
	QueryString []string `mapstructure:"input_query_strings"`
	// Endpoint Extra configuration for customized behaviour
	ExtraConfig ExtraConfig `mapstructure:"extra_config"`
	// HeadersToPass defines the list of headers to pass to the backends
	HeadersToPass []string `mapstructure:"input_headers"`
	// OutputEncoding defines the encoding strategy to use for the endpoint responses
	OutputEncoding string `mapstructure:"output_encoding"`
}

// Backend defines how lura should connect to the backend service (the API resource to consume)
// and how it should process the received response
type Backend struct {
	// Group defines the name of the property the response should be moved to. If empty, the response is
	// not changed
	Group string `mapstructure:"group"`
	// Method defines the HTTP method of the request to send to the backend
	Method string `mapstructure:"method"`
	// Host is a set of hosts of the API
	Host []string `mapstructure:"host"`
	// HostSanitizationDisabled can be set to false if the hostname should be sanitized
	HostSanitizationDisabled bool `mapstructure:"disable_host_sanitize"`
	// URLPattern is the URL pattern to use to locate the resource to be consumed
	URLPattern string `mapstructure:"url_pattern"`
	// AllowList is a set of response fields to allow. If empty, the filter id not used
	AllowList []string `mapstructure:"allow"`
	// DenyList is a set of response fields to remove. If empty, the filter id not used
	DenyList []string `mapstructure:"deny"`
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
	Decoder encoding.Decoder `json:"-"`
	// Backend Extra configuration for customized behaviours
	ExtraConfig ExtraConfig `mapstructure:"extra_config"`
}

// Plugin contains the config required by the plugin module
type Plugin struct {
	Folder  string `mapstructure:"folder"`
	Pattern string `mapstructure:"pattern"`
}

// TLS defines the configuration params for enabling TLS (HTTPS & HTTP/2) at the router layer
type TLS struct {
	IsDisabled               bool     `mapstructure:"disabled"`
	PublicKey                string   `mapstructure:"public_key"`
	PrivateKey               string   `mapstructure:"private_key"`
	MinVersion               string   `mapstructure:"min_version"`
	MaxVersion               string   `mapstructure:"max_version"`
	CurvePreferences         []uint16 `mapstructure:"curve_preferences"`
	PreferServerCipherSuites bool     `mapstructure:"prefer_server_cipher_suites"`
	CipherSuites             []uint16 `mapstructure:"cipher_suites"`
	EnableMTLS               bool     `mapstructure:"enable_mtls"`
}

// ExtraConfig is a type to store extra configurations for customized behaviours
type ExtraConfig map[string]interface{}

func (e *ExtraConfig) sanitize() {
	for module, extra := range *e {
		if extra, ok := extra.(map[interface{}]interface{}); ok {
			sanitized := map[string]interface{}{}
			for k, v := range extra {
				sanitized[fmt.Sprintf("%v", k)] = v
			}
			(*e)[module] = sanitized
		}
	}
}

func (e *ExtraConfig) Normalize() {
	for module := range *e {
		if alias, ok := ExtraConfigAlias[module]; ok {
			(*e)[alias] = (*e)[module]
			delete(*e, module)
		}
	}
}

// ExtraConfigAlias is the set of alias to accept as namespace
var ExtraConfigAlias = map[string]string{}

var (
	simpleURLKeysPattern    = regexp.MustCompile(`\{([\w\-\.:/]+)\}`)
	sequentialParamsPattern = regexp.MustCompile(`^(resp[\d]+_.+)?(JWT\.([\w\-\.:/]+))?$`)
	debugPattern            = "^[^/]|/__debug(/.*)?$"
	errInvalidHost          = errors.New("invalid host")
	errInvalidNoOpEncoding  = errors.New("can not use NoOp encoding with more than one backends connected to the same endpoint")
	defaultPort             = 8080
)

// Hash returns the sha 256 hash of the configuration in a standard base64 encoded string. It ignores the
// name in order to reduce the noise.
func (s *ServiceConfig) Hash() (string, error) {
	var name string
	name, s.Name = s.Name, ""
	defer func() { s.Name = name }()

	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return base64.StdEncoding.EncodeToString(sum[:]), nil
}

// Init initializes the configuration struct and its defined endpoints and backends.
// Init also sanitizes the values, applies the default ones whenever necessary and
// normalizes all the things.
func (s *ServiceConfig) Init() error {
	s.uriParser = NewURIParser()

	if s.Version != ConfigVersion {
		return &UnsupportedVersionError{
			Have: s.Version,
			Want: ConfigVersion,
		}
	}

	s.initGlobalParams()

	s.initAsyncAgents()

	return s.initEndpoints()
}

func (s *ServiceConfig) Normalize() {
	s.ExtraConfig.Normalize()
	for _, e := range s.Endpoints {
		e.ExtraConfig.Normalize()
		for _, b := range e.Backend {
			b.ExtraConfig.Normalize()
		}
	}
	for _, a := range s.AsyncAgents {
		a.ExtraConfig.Normalize()
		for _, b := range a.Backend {
			b.ExtraConfig.Normalize()
		}
	}
}

func (s *ServiceConfig) initGlobalParams() {
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

	s.ExtraConfig.sanitize()
}

func (s *ServiceConfig) initAsyncAgents() error {
	for i, e := range s.AsyncAgents {
		s.initAsyncAgentDefaults(i)

		e.ExtraConfig.sanitize()

		for _, b := range e.Backend {
			if len(b.Host) == 0 {
				b.Host = s.Host
			} else if !b.HostSanitizationDisabled {
				b.Host = s.uriParser.CleanHosts(b.Host)
			}
			if b.Method == "" {
				b.Method = http.MethodGet
			}
			b.Timeout = e.Consumer.Timeout
			b.Decoder = encoding.GetRegister().Get(strings.ToLower(b.Encoding))(b.IsCollection)

			b.ExtraConfig.sanitize()
		}
	}
	return nil
}

func (s *ServiceConfig) initEndpoints() error {
	for i, e := range s.Endpoints {
		e.Endpoint = s.uriParser.CleanPath(e.Endpoint)

		if err := e.validate(); err != nil {
			return err
		}

		for i := range e.HeadersToPass {
			e.HeadersToPass[i] = textproto.CanonicalMIMEHeaderKey(e.HeadersToPass[i])
		}

		inputParams := s.extractPlaceHoldersFromURLTemplate(e.Endpoint, s.paramExtractionPattern())
		inputSet := map[string]interface{}{}
		for ip := range inputParams {
			inputSet[inputParams[ip]] = nil
		}

		e.Endpoint = s.uriParser.GetEndpointPath(e.Endpoint, inputParams)

		s.initEndpointDefaults(i)

		if e.OutputEncoding == encoding.NOOP && len(e.Backend) > 1 {
			return errInvalidNoOpEncoding
		}

		e.ExtraConfig.sanitize()

		for j, b := range e.Backend {
			s.initBackendDefaults(i, j)

			if err := s.initBackendURLMappings(i, j, inputSet); err != nil {
				return err
			}

			b.ExtraConfig.sanitize()
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

func (*ServiceConfig) extractPlaceHoldersFromURLTemplate(subject string, pattern *regexp.Regexp) []string {
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
	if endpoint.OutputEncoding == "" {
		if s.OutputEncoding != "" {
			endpoint.OutputEncoding = s.OutputEncoding
		} else {
			endpoint.OutputEncoding = encoding.JSON
		}
	}
}

func (s *ServiceConfig) initAsyncAgentDefaults(e int) {
	agent := s.AsyncAgents[e]
	if s.Timeout != 0 && agent.Consumer.Timeout == 0 {
		agent.Consumer.Timeout = s.Timeout
	}
	if agent.Consumer.Workers < 1 {
		agent.Consumer.Workers = 1
	}
	if agent.Connection.HealthInterval < time.Second {
		agent.Connection.HealthInterval = time.Second
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
	if endpoint.OutputEncoding == encoding.NOOP {
		backend.Encoding = encoding.NOOP
	}
	backend.Timeout = endpoint.Timeout
	backend.ConcurrentCalls = endpoint.ConcurrentCalls
	backend.Decoder = encoding.GetRegister().Get(strings.ToLower(backend.Encoding))(backend.IsCollection)
}

func (s *ServiceConfig) initBackendURLMappings(e, b int, inputParams map[string]interface{}) error {
	backend := s.Endpoints[e].Backend[b]

	backend.URLPattern = s.uriParser.CleanPath(backend.URLPattern)

	outputParams, outputSetSize := uniqueOutput(s.extractPlaceHoldersFromURLTemplate(backend.URLPattern, simpleURLKeysPattern))

	ip := fromSetToSortedSlice(inputParams)

	if outputSetSize > len(ip) {
		return &WrongNumberOfParamsError{
			Endpoint:     s.Endpoints[e].Endpoint,
			Method:       s.Endpoints[e].Method,
			Backend:      b,
			InputParams:  ip,
			OutputParams: outputParams,
		}
	}

	backend.URLKeys = []string{}
	for _, output := range outputParams {
		if !sequentialParamsPattern.MatchString(output) {
			if _, ok := inputParams[output]; !ok {
				return &UndefinedOutputParamError{
					Param:        output,
					Endpoint:     s.Endpoints[e].Endpoint,
					Method:       s.Endpoints[e].Method,
					Backend:      b,
					InputParams:  ip,
					OutputParams: outputParams,
				}
			}
		}
		key := strings.Title(output[:1]) + output[1:]
		backend.URLPattern = strings.ReplaceAll(backend.URLPattern, "{"+output+"}", "{{."+key+"}}")
		backend.URLKeys = append(backend.URLKeys, key)
	}
	return nil
}

func fromSetToSortedSlice(set map[string]interface{}) []string {
	res := make([]string, 0, len(set))
	for element := range set {
		res = append(res, element)
	}
	sort.Strings(res)
	return res
}

func uniqueOutput(output []string) ([]string, int) {
	sort.Strings(output)
	j := 0
	outputSetSize := 0
	for i := 1; i < len(output); i++ {
		if output[j] == output[i] {
			continue
		}
		if !sequentialParamsPattern.MatchString(output[j]) {
			outputSetSize++
		}
		j++
		output[j] = output[i]
	}
	if j == len(output) {
		return output, outputSetSize
	}
	return output[:j+1], outputSetSize
}

func (e *EndpointConfig) validate() error {
	matched, err := regexp.MatchString(debugPattern, e.Endpoint)
	if err != nil {
		return &EndpointMatchError{
			Err:    err,
			Path:   e.Endpoint,
			Method: e.Method,
		}
	}
	if matched {
		return &EndpointPathError{Path: e.Endpoint, Method: e.Method}
	}

	if len(e.Backend) == 0 {
		return &NoBackendsError{Path: e.Endpoint, Method: e.Method}
	}
	return nil
}

// EndpointMatchError is the error returned by the configuration init process when the endpoint pattern
// check fails
type EndpointMatchError struct {
	Path   string
	Method string
	Err    error
}

// Error returns a string representation of the EndpointMatchError
func (e *EndpointMatchError) Error() string {
	return fmt.Sprintf("ignoring the '%s %s' endpoint due to a parsing error: %s", e.Method, e.Path, e.Err.Error())
}

// NoBackendsError is the error returned by the configuration init process when an endpoint
// is connected to 0 backends
type NoBackendsError struct {
	Path   string
	Method string
}

// Error returns a string representation of the NoBackendsError
func (n *NoBackendsError) Error() string {
	return "ignoring the '" + n.Method + " " + n.Path + "' endpoint, since it has 0 backends defined!"
}

// UnsupportedVersionError is the error returned by the configuration init process when the configuration
// version is not supoprted
type UnsupportedVersionError struct {
	Have int
	Want int
}

// Error returns a string representation of the UnsupportedVersionError
func (u *UnsupportedVersionError) Error() string {
	return fmt.Sprintf("unsupported version: %d (want: %d)", u.Have, u.Want)
}

// EndpointPathError is the error returned by the configuration init process when an endpoint
// is using a forbidden path
type EndpointPathError struct {
	Path   string
	Method string
}

// Error returns a string representation of the EndpointPathError
func (e *EndpointPathError) Error() string {
	return "ignoring the '" + e.Method + " " + e.Path + "' endpoint, since it is invalid!!!"
}

// UndefinedOutputParamError is the error returned by the configuration init process when an output
// param is not present in the input param set
type UndefinedOutputParamError struct {
	Endpoint     string
	Method       string
	Backend      int
	InputParams  []string
	OutputParams []string
	Param        string
}

// Error returns a string representation of the UndefinedOutputParamError
func (u *UndefinedOutputParamError) Error() string {
	return fmt.Sprintf(
		"undefined output param '%s'! endpoint: %s %s, backend: %d. input: %v, output: %v",
		u.Param,
		u.Method,
		u.Endpoint,
		u.Backend,
		u.InputParams,
		u.OutputParams,
	)
}

// WrongNumberOfParamsError is the error returned by the configuration init process when the number of output
// params is greatter than the number of input params
type WrongNumberOfParamsError struct {
	Endpoint     string
	Method       string
	Backend      int
	InputParams  []string
	OutputParams []string
}

// Error returns a string representation of the WrongNumberOfParamsError
func (w *WrongNumberOfParamsError) Error() string {
	return fmt.Sprintf(
		"input and output params do not match. endpoint: %s %s, backend: %d. input: %v, output: %v",
		w.Method,
		w.Endpoint,
		w.Backend,
		w.InputParams,
		w.OutputParams,
	)
}
