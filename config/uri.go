// SPDX-License-Identifier: Apache-2.0

package config

import (
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

// NewURIParser creates a new URIParser using the package variable RoutingPattern
func NewURIParser() URIParser {
	return URI(RoutingPattern)
}

// URI implements the URIParser interface
type URI int

// CleanHosts applies the CleanHost method to every member of the received array of hosts
func (u URI) CleanHosts(hosts []string) []string {
	cleaned := make([]string, 0, len(hosts))
	for i := range hosts {
		cleaned = append(cleaned, u.CleanHost(hosts[i]))
	}
	return cleaned
}

// CleanHost sanitizes the received host
func (URI) CleanHost(host string) string {
	matches := hostPattern.FindAllStringSubmatch(host, -1)
	if len(matches) != 1 {
		panic(errInvalidHost)
	}
	keys := matches[0][1:]
	if keys[0] == "" {
		keys[0] = "http://"
	}
	return strings.Join(keys, "")
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
			parts[0] = strings.ReplaceAll(parts[0], "{"+params[p]+"}", ":"+params[p])
			result = strings.Join(parts, "?")
		}
	}
	return result
}
