// SPDX-License-Identifier: Apache-2.0
package mux

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/luraproject/lura/logging"
)

// DebugHandler creates a dummy handler function, useful for quick integration tests
func DebugHandler(logger logging.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Method:", r.Method)
		logger.Debug("URL:", r.RequestURI)
		logger.Debug("Query:", r.URL.Query())
		// logger.Debug("Params:", c.Params)
		logger.Debug("Headers:", r.Header)
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		logger.Debug("Body:", string(body))

		js, _ := json.Marshal(map[string]string{"message": "pong"})

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
