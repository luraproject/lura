//go:build integration || !race
// +build integration !race

// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

func TestProxyStack_multi(t *testing.T) {
	results := map[string]int{}
	m := new(sync.Mutex)
	total := 100000
	cfgPath := ".config.json"

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.Lock()
		results[r.URL.String()]++
		m.Unlock()
		w.Write([]byte("{\"foo\":42}"))
	}))
	defer s.Close()

	{
		cfgContent := `{
			"version":3,
			"endpoints":[{
				"endpoint":"/{foo}",
				"backend":[
					{
						"host":       ["%s"],
						"url_pattern": "/first/{foo}",
						"group":      "1"
					},
					{
						"host":       ["%s"],
						"url_pattern": "/second/{foo}",
						"group":      "2"
					},
					{
						"host":       ["%s"],
						"url_pattern": "/third/{foo}",
						"group":      "3"
					}
				]
			}]
		}`
		if err := os.WriteFile(cfgPath, []byte(fmt.Sprintf(cfgContent, s.URL, s.URL, s.URL)), 0666); err != nil {
			log.Fatal(err)
		}
		defer os.Remove(cfgPath)
	}

	cfg, err := config.NewParser().Parse(cfgPath)
	if err != nil {
		t.Error(err)
		return
	}
	cfg.Normalize()

	factory := NewDefaultFactory(httpProxy, logging.NoOp)
	p, err := factory.New(cfg.Endpoints[0])
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < total; i++ {
		p(context.Background(), &Request{
			Method:  "GET",
			Params:  map[string]string{"Foo": "42"},
			Headers: map[string][]string{},
			Path:    "/",
			Query:   url.Values{},
			Body:    io.NopCloser(strings.NewReader("")),
			URL:     new(url.URL),
		})
	}

	for k, v := range results {
		if v != total {
			t.Errorf("the url %s was consumed %d times", k, v)
		}
	}

	if len(results) != 3 {
		t.Errorf("unexpected number of consumed urls. have %d, want 3", len(results))
	}
}
