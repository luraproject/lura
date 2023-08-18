// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestNewFilterHeadersMiddleware(t *testing.T) {
	mw := NewFilterHeadersMiddleware(
		logging.NoOp,
		&config.Backend{
			HeadersToPass: []string{
				"X-This-Shall-Pass",
				"X-Gandalf-Will-Pass",
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
		Headers: map[string][]string{
			"X-This-Shall-Pass":    []string{"tupu", "supu"},
			"X-You-Shall-Not-Pass": []string{"Balrog"},
			"X-Gandalf-Will-Pass":  []string{"White", "Grey"},
			"X-Drop-Tables":        []string{"foo"},
		},
	}

	prxy(context.Background(), sentReq)

	if receivedReq == sentReq {
		t.Errorf("request should be different")
		return
	}

	if _, ok := receivedReq.Headers["X-This-Shall-Pass"]; !ok {
		t.Errorf("missing X-This-Shall-Pass")
		return
	}

	if _, ok := receivedReq.Headers["X-Gandalf-Will-Pass"]; !ok {
		t.Errorf("missing X-Gandalf-Will-Pass")
		return
	}

	if _, ok := receivedReq.Headers["X-Drop-Tables"]; ok {
		t.Errorf("should not be there X-Drop-Tables")
		return
	}

	if _, ok := receivedReq.Headers["X-You-Shall-Not-Pass"]; ok {
		t.Errorf("should not be there X-You-Shall-Not-Pass")
		return
	}

	// check that when headers are the expected, no need to copy
	sentReq = &Request{
		Body:   nil,
		Params: map[string]string{},
		Headers: map[string][]string{
			"X-This-Shall-Pass": []string{"tupu", "supu"},
		},
	}

	prxy(context.Background(), sentReq)

	if receivedReq != sentReq {
		t.Errorf("request should be the same, no modification of headers expected")
		return
	}
}
