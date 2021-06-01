// +build integration !race

// SPDX-License-Identifier: Apache-2.0
package proxy

import (
	"context"
	"fmt"
	"testing"

	"github.com/luraproject/lura/config"
	"github.com/luraproject/lura/proxy/plugin"
)

func TestNewPluginMiddleware(t *testing.T) {
	plugin.LoadModifiers("./plugin/tests", ".so", plugin.RegisterModifier)

	validator := func(ctx context.Context, r *Request) (*Response, error) {
		if r.Path != "/bar/fooo/fooo" {
			return nil, fmt.Errorf("unexpected path %s", r.Path)
		}
		return nil, nil
	}

	bknd := NewBackendPluginMiddleware(
		&config.Backend{
			ExtraConfig: map[string]interface{}{
				plugin.Namespace: map[string]interface{}{
					"name": []interface{}{"lura-request-modifier-example"},
				},
			},
		},
	)(validator)

	p := NewPluginMiddleware(
		&config.EndpointConfig{
			ExtraConfig: map[string]interface{}{
				plugin.Namespace: map[string]interface{}{
					"name": []interface{}{"lura-request-modifier-example"},
				},
			},
		},
	)(bknd)

	resp, err := p(context.Background(), &Request{Path: "/bar"})
	if err != nil {
		t.Error(err.Error())
	}

	if resp != nil {
		t.Errorf("unexpected response: %v", resp)
	}
}
