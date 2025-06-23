// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"
)

// ResponseWrapper defines the interface to access the inner
// http.ResponseWriter (for example, gin's ResponseWriter
// wraps a http.ResponseWriter).
type ResponseWrapper interface {
	Unwrap() http.ResponseWriter
}

// WrappedHTTPFlusher checks if there is a inner http.ResponseWriter
// that implements http.Flusher, and returns it directly, if not,
// check if the writer implements it, and returns it, and otherwise
// it returns nil.
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
