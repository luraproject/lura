package router

import (
	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
	"testing"
)

func TestIsValidSequentialEndpoint_ok(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "GET",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": true,
			},
		},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if !success {
		t.Error("Endpoint expected valid but receive invalid")
	}
}

func TestIsValidSequentialEndpoint_wrong_config_not_given(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "GET",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}

func TestIsValidSequentialEndpoint_wrong_config_set_false(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "GET",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": false,
			},
		}}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}

func TestIsValidSequentialEndpoint_wrong_order(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "PUT",
			},
			{
				Method: "GET",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": true,
			},
		},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}

func TestIsValidSequentialEndpoint_wrong_all_non_get(t *testing.T) {

	endpoint := &config.EndpointConfig{
		Endpoint: "/correct",
		Method:   "PUT",
		Backend: []*config.Backend{
			{
				Method: "POST",
			},
			{
				Method: "PUT",
			},
		},
		ExtraConfig: map[string]interface{}{
			proxy.Namespace: map[string]interface{}{
				"sequential": true,
			},
		},
	}

	success := IsValidSequentialEndpoint(endpoint)

	if success {
		t.Error("Endpoint expected invalid but receive valid")
	}
}
