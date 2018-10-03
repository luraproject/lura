package proxy

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/devopsfaith/krakend/config"
)

func TestDetailedHTTPStatusHandler(t *testing.T) {
	cfg := &config.Backend{
		ExtraConfig: config.ExtraConfig{
			Namespace: map[string]interface{}{
				"return_error_details": true,
			},
		},
	}
	sh := getHTTPStatusHandler(cfg)

	for i, code := range statusCodes {
		msg := http.StatusText(code)

		resp := &http.Response{
			StatusCode: code,
			Body:       ioutil.NopCloser(bytes.NewBufferString(msg)),
		}

		r, err := sh(context.Background(), resp)

		if r != resp {
			t.Errorf("#%d unexpected response: %v", i, r)
			return
		}

		e, ok := err.(HTTPResponseError)
		if !ok {
			t.Errorf("#%d unexpected error type %T: %s", i, err, err.Error())
			return
		}

		if e.Code != code {
			t.Errorf("#%d unexpected status code: %d", i, e.Code)
			return
		}

		if e.Msg != msg {
			t.Errorf("#%d unexpected message: %s", i, e.Msg)
			return
		}

		expectedResponse := &Response{
			Data: map[string]interface{}{
				"error": HTTPResponseError{
					Code: code,
					Msg:  msg,
				},
			},
			Metadata: Metadata{StatusCode: code},
		}

		if !reflect.DeepEqual(e.Response(), expectedResponse) {
			t.Errorf("#%d unexpected response: %v", i, e.Response())
		}
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
