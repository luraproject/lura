// +build integration !race

// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestLoadWithLogger(t *testing.T) {
	buff := new(bytes.Buffer)
	l, _ := logging.NewLogger("DEBUG", buff, "")
	total, err := LoadWithLogger("./tests", ".so", RegisterHandler, l)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if total != 1 {
		t.Errorf("unexpected number of loaded plugins!. have %d, want 1", total)
	}

	var handler http.Handler

	hre := New(l, func(_ context.Context, _ config.ServiceConfig, h http.Handler) error {
		handler = h
		return nil
	})

	if err := hre(
		context.Background(),
		config.ServiceConfig{
			ExtraConfig: map[string]interface{}{
				Namespace: map[string]interface{}{
					"name": "krakend-server-example",
				},
			},
		},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("this handler should not been called")
		}),
	); err != nil {
		t.Error(err)
		return
	}

	req, _ := http.NewRequest("GET", "http://some.example.tld/path", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	if resp.StatusCode != 200 {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return
	}
	resp.Body.Close()

	if string(b) != "Hello, \"/path\"" {
		t.Errorf("unexpected response body: %s", string(b))
	}

	fmt.Println(buff.String())
}
