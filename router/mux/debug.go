package mux

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/devopsfaith/krakend/logging"
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

		js, err := json.Marshal(map[string]string{"message": "pong"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}
