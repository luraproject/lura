// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"
)

// gin implements unwrap to get the inner response writer
type ResponseWrapper interface {
	Unwrap() http.ResponseWriter
}

func WrappedHTTPFlusher(w http.ResponseWriter) http.Flusher {
	unwrapper, ok := w.(ResponseWrapper)
	if ok {
		ww := unwrapper.Unwrap()
		if ww != nil {
			w = ww
		}
	}

	f, ok := w.(http.Flusher)
	if !ok {
		return nil
	}
	return f
}
