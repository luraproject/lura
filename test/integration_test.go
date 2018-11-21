// +build integration !race

package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/logging"
	"github.com/devopsfaith/krakend/proxy"
	"github.com/devopsfaith/krakend/router/gin"
)

func TestKrakenD(t *testing.T) {
	cfg, err := setupBackend(t)
	if err != nil {
		t.Error(err)
		return
	}

	buf := new(bytes.Buffer)
	logger, err := logging.NewLogger("DEBUG", buf, "[KRAKEND]")
	if err != nil {
		t.Error(err)
		return
	}

	go func() {
		gin.DefaultFactory(proxy.DefaultFactory(logger), logger).New().Run(*cfg)
	}()

	for _, tc := range []struct {
		name    string
		url     string
		method  string
		headers map[string]string
		body    string
		expBody string
	}{
		{
			name:    "static",
			url:     "/static",
			headers: map[string]string{},
			expBody: `{"bar":"foobar","foo":42}`,
		},
		{
			name:   "param_forwarding",
			url:    "/param_forwarding/foo",
			method: "POST",
			headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "bearer AuthorizationToken",
				"X-Y-Z":         "x-y-z",
			},
			body:    `{"foo":"bar"}`,
			expBody: `{"path":"/foo"}`,
		},
		{
			name:    "timeout",
			url:     "/timeout",
			headers: map[string]string{},
			expBody: `{"email":"some@email.com","name":"a"}`,
		},
		{
			name:    "partial_with_static",
			url:     "/partial/static",
			headers: map[string]string{},
			expBody: `{"bar":"foobar","email":"some@email.com","foo":42,"name":"a"}`,
		},
		{
			name:    "partial",
			url:     "/partial",
			headers: map[string]string{},
			expBody: `{"email":"some@email.com","name":"a"}`,
		},
		{
			name:    "combination",
			url:     "/combination",
			headers: map[string]string{},
			expBody: `{"name":"a","personal_email":"some@email.com","posts":[{"body":"some content","date":"123456789"},{"body":"some other content","date":"123496789"}]}`,
		},
		{
			name:    "detail_error",
			url:     "/detail_error",
			headers: map[string]string{},
			expBody: `{"email":"some@email.com","error_backend_a":{"http_status_code":429,"http_body":"sad panda\n"},"name":"a"}`,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			time.Sleep(300 * time.Millisecond)

			if tc.method == "" {
				tc.method = "GET"
			}

			var body io.Reader
			if tc.body != "" {
				body = bytes.NewBufferString(tc.body)
			}

			r, _ := http.NewRequest(tc.method, fmt.Sprintf("http://localhost:%d%s", cfg.Port, tc.url), body)
			for k, v := range tc.headers {
				r.Header.Add(k, v)
			}

			resp, err := http.DefaultClient.Do(r)
			if err != nil {
				t.Error(err)
				return
			}
			if resp == nil {
				t.Errorf("%s: nil response", resp.Request.URL.Path)
				return
			}

			if c := resp.Header.Get("Content-Type"); c != "application/json; charset=utf-8" {
				t.Errorf("%s: unexpected header content-type: %s", resp.Request.URL.Path, c)
				return
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("%s: unexpected status code: %d", resp.Request.URL.Path, resp.StatusCode)
				return
			}
			b, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if tc.expBody != string(b) {
				t.Errorf("%s: unexpected body: %s", resp.Request.URL.Path, string(b))
				return
			}
		})
	}

}

func setupBackend(t *testing.T) (*config.ServiceConfig, error) {
	data := map[string]interface{}{"port": rand.Intn(2000) + 8080}

	// param forwarding validation backend
	b1 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if c := r.Header.Get("Content-Type"); c != "application/json" {
			t.Errorf("unexpected header content-type: %s", c)
			http.Error(rw, "bad content-type", 400)
			return
		}
		if c := r.Header.Get("Authorization"); c != "bearer AuthorizationToken" {
			t.Errorf("unexpected header Authorization: %s", c)
			http.Error(rw, "bad Authorization", 400)
			return
		}
		if c := r.Header.Get("X-Y-Z"); c != "x-y-z" {
			t.Errorf("unexpected header X-Y-Z: %s", c)
			http.Error(rw, "bad X-Y-Z", 400)
			return
		}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		if string(body) != `{"foo":"bar"}` {
			t.Errorf("unexpected request body: %s", string(body))
			return
		}
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{"path": r.URL.Path})
	}))
	data["b1"] = b1.URL

	// collection generator backend
	b2 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode([]interface{}{
			map[string]interface{}{"body": "some content", "date": "123456789"},
			map[string]interface{}{"body": "some other content", "date": "123496789"},
		})
	}))
	data["b2"] = b2.URL

	// regular struct generator backend
	b3 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{"email": "some@email.com", "name": "a"})
	}))
	data["b3"] = b3.URL

	// crasher backend
	b4 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		http.Error(rw, "sad panda", 429)
	}))
	data["b4"] = b4.URL

	// slow backend
	b5 := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		<-time.After(time.Second)
		rw.Header().Add("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(map[string]interface{}{"email": "some@email.com", "name": "a"})
	}))
	data["b5"] = b5.URL

	c, err := loadConfig(data)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func loadConfig(data map[string]interface{}) (*config.ServiceConfig, error) {
	content, _ := ioutil.ReadFile("krakend.json")
	tmpl, err := template.New("test").Parse(string(content))
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, data); err != nil {
		return nil, err
	}

	c, err := config.NewParserWithFileReader(func(s string) ([]byte, error) {
		return []byte(s), nil
	}).Parse(buf.String())

	if err != nil {
		return nil, err
	}

	return &c, nil
}
