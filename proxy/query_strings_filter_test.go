// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestNewFilterQueryStringsMiddleware(t *testing.T) {
	mw := NewFilterQueryStringsMiddleware(
		logging.NoOp,
		&config.Backend{
			QueryStringsToPass: []string{
				"oak",
				"cedar",
			},
		},
	)

	var receivedReq *Request
	prxy := mw(func(ctx context.Context, req *Request) (*Response, error) {
		receivedReq = req
		return nil, nil
	})

	sentReq := &Request{
		Body:   nil,
		Params: map[string]string{},
		Query: map[string][]string{
			"oak":    []string{"acorn", "evergreen"},
			"maple":  []string{"tree", "shrub"},
			"cedar":  []string{"mediterranean", "himalayas"},
			"willow": []string{"350"},
		},
	}

	prxy(context.Background(), sentReq)

	if receivedReq == sentReq {
		t.Errorf("request should be different")
		return
	}

	oak, ok := receivedReq.Query["oak"]
	if !ok {
		t.Errorf("missing 'oak'")
		return
	}
	if len(oak) != len(sentReq.Query["oak"]) {
		t.Errorf("want len(oak): %d, got %d",
			len(sentReq.Query["oak"]), len(oak))
		return
	}

	for idx, expected := range sentReq.Query["oak"] {
		if expected != oak[idx] {
			t.Errorf("want oak[%d] = %s, got %s",
				idx, expected, oak[idx])
			return
		}
	}

	if _, ok := receivedReq.Query["cedar"]; !ok {
		t.Errorf("missing 'cedar'")
		return
	}

	if _, ok := receivedReq.Query["mapple"]; ok {
		t.Errorf("should not be there: 'mapple'")
		return
	}

	if _, ok := receivedReq.Query["willow"]; ok {
		t.Errorf("should not be there: 'willow'")
		return
	}

	// check that when query strings are all the expected, no need to copy
	sentReq = &Request{
		Body:   nil,
		Params: map[string]string{},
		Query: map[string][]string{
			"oak":   []string{"acorn", "evergreen"},
			"cedar": []string{"mediterranean", "himalayas"},
		},
	}

	prxy(context.Background(), sentReq)

	if receivedReq != sentReq {
		t.Errorf("request should be the same, no modification of query string expected")
		return
	}
}

func TestFilterQueryStringsBlockAll(t *testing.T) {
	// In order to block all the query strings, we must only let pass
	// the 'empty' string ""
	mw := NewFilterQueryStringsMiddleware(
		logging.NoOp,
		&config.Backend{
			QueryStringsToPass: []string{""},
		},
	)

	var receivedReq *Request
	prxy := mw(func(ctx context.Context, req *Request) (*Response, error) {
		receivedReq = req
		return nil, nil
	})

	sentReq := &Request{
		Body:   nil,
		Params: map[string]string{},
		Query: map[string][]string{
			"oak":   []string{"acorn", "evergreen"},
			"maple": []string{"tree", "shrub"},
		},
	}

	prxy(context.Background(), sentReq)

	if receivedReq == sentReq {
		t.Errorf("request should be different")
		return
	}

	if len(receivedReq.Query) != 0 {
		t.Errorf("should have blocked all query strings")
		return
	}
}

func TestFilterQueryStringsAllowAll(t *testing.T) {
	// Empty backend query strings to passa everything
	mw := NewFilterQueryStringsMiddleware(
		logging.NoOp,
		&config.Backend{
			QueryStringsToPass: []string{},
		},
	)

	var receivedReq *Request
	prxy := mw(func(ctx context.Context, req *Request) (*Response, error) {
		receivedReq = req
		return nil, nil
	})

	sentReq := &Request{
		Body:   nil,
		Params: map[string]string{},
		Query: map[string][]string{
			"oak":   []string{"acorn", "evergreen"},
			"maple": []string{"tree", "shrub"},
		},
	}

	prxy(context.Background(), sentReq)

	if len(receivedReq.Query) != 2 {
		t.Errorf("should have passed all query strings")
		return
	}
}
