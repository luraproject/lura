// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	endpointURLKeysPattern = regexp.MustCompile(`/\{([a-zA-Z\-_0-9]+)\}`)
	hostPattern            = regexp.MustCompile(`(https?://)?([a-zA-Z0-9\._\-]+)(:[0-9]{2,6})?/?`)
)

// URIParser defines the interface for all the URI manipulation required by KrakenD
type URIParser interface {
	CleanHosts([]string) []string
	CleanHost(string) string
	CleanPath(string) string
	GetEndpointPath(string, []string) string
}

// Like URIParser but with safe versions of the clean host functionality that
// does not panic but returns an error.
type SafeURIParser interface {
	SafeCleanHosts([]string) ([]string, error)
	SafeCleanHost(string) (string, error)
	CleanPath(string) string
	GetEndpointPath(string, []string) string
}

// NewURIParser creates a new URIParser using the package variable RoutingPattern
func NewURIParser() URIParser {
	return URI(RoutingPattern)
}

// NewSafeURIParser creates a safe URI parser that does not panic when cleaning hosts
func NewSafeURIParser() URI {
	return URI(RoutingPattern)
}

// URI implements the URIParser interface
type URI int

// SafeCleanHosts applies the SafeCleanHost method to every member of the received array of hosts
func (u URI) SafeCleanHosts(hosts []string) ([]string, error) {
	cleaned := make([]string, 0, len(hosts))
	for i := range hosts {
		h, err := u.SafeCleanHost(hosts[i])
		if err != nil {
			return nil, fmt.Errorf("host %s not valid: %w", hosts[i], errInvalidHost)
		}
		cleaned = append(cleaned, h)
	}
	return cleaned, nil
}

// CleanHosts applies the CleanHost method to every member of the received array of hosts
// Panics in case of error.
func (u URI) CleanHosts(hosts []string) []string {
	ss, e := u.SafeCleanHosts(hosts)
	if e != nil {
		panic(e)
	}
	return ss
}

// SafeCleanHost sanitizes the received host
func (URI) SafeCleanHost(host string) (string, error) {
	matches := hostPattern.FindAllStringSubmatch(host, -1)
	if len(matches) != 1 {
		return "", errInvalidHost
	}
	keys := matches[0][1:]
	if keys[0] == "" {
		keys[0] = "http://"
	}
	return strings.Join(keys, ""), nil
}

// CleanHost sanitizes the received host.
// Panics on error.
func (u URI) CleanHost(host string) string {
	h, err := u.SafeCleanHost(host)
	if err != nil {
		panic(err)
	}
	return h
}

// CleanPath trims all the extra slashes from the received URI path
func (URI) CleanPath(path string) string {
	return "/" + strings.TrimPrefix(path, "/")
}

// GetEndpointPath applies the proper replacement in the received path to generate valid route patterns
func (u URI) GetEndpointPath(path string, params []string) string {
	result := path
	if u == ColonRouterPatternBuilder {
		for p := range params {
			parts := strings.Split(result, "?")
			param := params[p]
			if !strings.HasPrefix(params[p], "*") {
				param = ":" + params[p]
			}
			parts[0] = strings.ReplaceAll(parts[0], "{"+params[p]+"}", param)
			result = strings.Join(parts, "?")
		}
	}
	return result
}
