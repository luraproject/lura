// SPDX-License-Identifier: Apache-2.0
package main

//
import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/http"

	"github.com/luraproject/lura/logging"
)

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the RegisterClient interface
var ClientRegisterer = registerer("krakend-client-example")

type registerer string

var logger logging.Logger = nil

func (r registerer) RegisterLogger(v interface{}) {
	l, ok := v.(logging.Logger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(ClientRegisterer, "client plugin loaded!!!")
}

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(ctx context.Context, extra map[string]interface{}) (http.Handler, error) {
	// check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	if logger == nil {
		// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http client
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(w, "Hello, %q", html.EscapeString(req.URL.Path))
		}), nil
	}

	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http client
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(req.URL.Path))
		logger.Debug("request:", html.EscapeString(req.URL.Path))
	}), nil
}

func init() {
	fmt.Println(ClientRegisterer, "client plugin loaded!!!")
}

func main() {}
