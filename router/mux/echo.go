// SPDX-License-Identifier: Apache-2.0

package mux

import (
	"encoding/json"
	"io"
	"net/http"
)

// DebugHandler creates a dummy handler function, useful for quick integration tests
func EchoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		js, _ := json.Marshal(map[string]interface{}{
			"Method":  r.Method,
			"URL":     r.RequestURI,
			"Query":   r.URL.Query(),
			"Headers": r.Header,
			"Body":    string(body),
		})

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
