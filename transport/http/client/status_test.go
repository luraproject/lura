// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/luraproject/lura/v2/config"
)

func TestDetailedHTTPStatusHandler(t *testing.T) {
	expectedErrName := "some"
	expectedEncoding := "application/json; charset=utf-8"
	cfg := &config.Backend{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				"return_error_details": expectedErrName,
			},
		},
	}
	sh := GetHTTPStatusHandler(cfg)

	for _, code := range []int{http.StatusOK, http.StatusCreated} {
		resp := &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`)),
		}

		r, err := sh(context.Background(), resp)

		if r != resp {
			t.Errorf("#%d unexpected response: %v", code, r)
			return
		}

		if err != nil {
			t.Errorf("#%d unexpected error: %s", code, err.Error())
			return
		}
	}

	for i, code := range statusCodes {
		msg := http.StatusText(code)

		resp := &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{"msg":%q}`, msg))),
			Header:     http.Header{"Content-Type": []string{expectedEncoding}},
		}

		r, err := sh(context.Background(), resp)

		if r != resp {
			t.Errorf("#%d unexpected response: %v", i, r)
			return
		}

		e, ok := err.(NamedHTTPResponseError)
		if !ok {
			t.Errorf("#%d unexpected error type %T: %s", i, err, err.Error())
			return
		}

		if e.StatusCode() != code {
			t.Errorf("#%d unexpected status code: %d", i, e.Code)
			return
		}

		if e.Error() != fmt.Sprintf(`{"msg":%q}`, msg) {
			t.Errorf("#%d unexpected message: %s", i, e.Msg)
			return
		}

		if e.Name() != expectedErrName {
			t.Errorf("#%d unexpected error name: %s", i, e.name)
			return
		}

		if e.Encoding() != expectedEncoding {
			t.Errorf("#%d unexpected encoding: %s", i, e.Enc)
		}
	}
}

func TestDefaultHTTPStatusHandler(t *testing.T) {
	sh := GetHTTPStatusHandler(&config.Backend{})

	for _, code := range []int{http.StatusOK, http.StatusCreated} {
		resp := &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`)),
		}

		r, err := sh(context.Background(), resp)

		if r != resp {
			t.Errorf("#%d unexpected response: %v", code, r)
			return
		}

		if err != nil {
			t.Errorf("#%d unexpected error: %s", code, err.Error())
			return
		}
	}

	for _, code := range statusCodes {
		msg := http.StatusText(code)

		resp := &http.Response{
			StatusCode: code,
			Body:       io.NopCloser(bytes.NewBufferString(msg)),
		}

		r, err := sh(context.Background(), resp)

		if r != nil {
			t.Errorf("#%d unexpected response: %v", code, r)
			return
		}

		if !strings.HasPrefix(err.Error(), "invalid status code") {
			t.Errorf("#%d unexpected error: %v", code, err)
			return
		}
	}
}

func TestNewHTTPResponseError_limitsBodySize(t *testing.T) {
	original := maxErrorResponseBody
	defer func() { maxErrorResponseBody = original }()
	maxErrorResponseBody = 8

	full := "0123456789abcdef" // 16 bytes, over the 8-byte cap
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Body:       io.NopCloser(strings.NewReader(full)),
	}

	respErr := newHTTPResponseError(resp)

	want := full[:maxErrorResponseBody]
	if respErr.Msg != want {
		t.Errorf("expected the error message truncated to %q, got %q", want, respErr.Msg)
		return
	}
	if respErr.Enc != "text/plain" {
		t.Errorf("expected the content type preserved, got %q", respErr.Enc)
		return
	}
	rest, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("re-reading the replaced body: %v", err)
		return
	}
	if string(rest) != want {
		t.Errorf("expected the replaced body to hold the truncated bytes %q, got %q", want, string(rest))
	}
}

var statusCodes = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusPaymentRequired,
	http.StatusForbidden,
	http.StatusNotFound,
	http.StatusMethodNotAllowed,
	http.StatusNotAcceptable,
	http.StatusProxyAuthRequired,
	http.StatusRequestTimeout,
	http.StatusConflict,
	http.StatusGone,
	http.StatusLengthRequired,
	http.StatusPreconditionFailed,
	http.StatusRequestEntityTooLarge,
	http.StatusRequestURITooLong,
	http.StatusUnsupportedMediaType,
	http.StatusRequestedRangeNotSatisfiable,
	http.StatusExpectationFailed,
	http.StatusTeapot,
	// http.StatusMisdirectedRequest,
	http.StatusUnprocessableEntity,
	http.StatusLocked,
	http.StatusFailedDependency,
	http.StatusUpgradeRequired,
	http.StatusPreconditionRequired,
	http.StatusTooManyRequests,
	http.StatusRequestHeaderFieldsTooLarge,
	http.StatusUnavailableForLegalReasons,

	http.StatusInternalServerError,
	http.StatusNotImplemented,
	http.StatusBadGateway,
	http.StatusServiceUnavailable,
	http.StatusGatewayTimeout,
	http.StatusHTTPVersionNotSupported,
	http.StatusVariantAlsoNegotiates,
	http.StatusInsufficientStorage,
	http.StatusLoopDetected,
	http.StatusNotExtended,
	http.StatusNetworkAuthenticationRequired,
}
