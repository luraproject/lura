// SPDX-License-Identifier: Apache-2.0

package mux

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/luraproject/lura/v2/logging"
)

// DebugHandler creates a dummy handler function, useful for quick integration tests
func DebugHandler(logger logging.Logger) http.HandlerFunc {
	logPrefixSecondary := "[ENDPOINT /__debug/*]"
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug(logPrefixSecondary, "Method:", r.Method)
		logger.Debug(logPrefixSecondary, "URL:", r.RequestURI)
		logger.Debug(logPrefixSecondary, "Query:", r.URL.Query())
		// logger.Debug(logPrefixSecondary, "Params:", c.Params)
		logger.Debug(logPrefixSecondary, "Headers:", r.Header)
		body, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		logger.Debug(logPrefixSecondary, "Body:", string(body))

		js, _ := json.Marshal(map[string]string{"message": "pong"})

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
