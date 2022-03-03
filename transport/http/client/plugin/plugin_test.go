// +build integration !race

// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/transport/http/client"
)

func TestLoadWithLogger(t *testing.T) {
	buff := new(bytes.Buffer)
	l, _ := logging.NewLogger("DEBUG", buff, "")
	total, err := LoadWithLogger("./tests", ".so", RegisterClient, l)
	if err != nil {
		t.Error(err.Error())
		t.Fail()
	}
	if total != 1 {
		t.Errorf("unexpected number of loaded plugins!. have %d, want 1", total)
	}

	hre := HTTPRequestExecutor(l, func(_ *config.Backend) client.HTTPRequestExecutor {
		t.Error("this factory should not been called")
		t.Fail()
		return nil
	})

	h := hre(&config.Backend{
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"name": "krakend-client-example",
			},
		},
	})

	req, _ := http.NewRequest("GET", "http://some.example.tld/path", nil)
	resp, err := h(context.Background(), req)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
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
