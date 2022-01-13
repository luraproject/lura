// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultHTTPRequestExecutor(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	re := DefaultHTTPRequestExecutor(NewHTTPClient)

	req, _ := http.NewRequest("GET", ts.URL, ioutil.NopCloser(&bytes.Buffer{}))

	resp, err := re(context.Background(), req)

	if err != nil {
		t.Error("unexpected error:", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("unexpected status code:", resp.StatusCode)
	}
}
