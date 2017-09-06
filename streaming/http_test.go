package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/proxy"
)

func TestNewHTTPProxy_ok(t *testing.T) {
	expectedMethod := "GET"
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != expectedMethod {
			t.Errorf("Wrong request method. Want: %s. Have: %s", expectedMethod, r.Method)
		}
		fmt.Fprintf(w, "{\"supu\":42, \"tupu\":true, \"foo\": \"bar\"}")
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{

	}
	request := proxy.Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	result, err := StreamHTTPProxyFactory(http.DefaultClient)(&backend)(context.Background(), &request)
	if err != nil {
		t.Errorf("The proxy returned an unexpected error: %s\n", err.Error())
		return
	}
	if result == nil {
		t.Errorf("The proxy returned a null result\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("Error: expected response")
		return
	default:
	}

	data := make(map[string]interface{})
	decoder := json.NewDecoder(result.Io)
	decoder.UseNumber()
	err = decoder.Decode(&data)
	if err != nil {
		t.Errorf("The proxy returned an undecodeable result\n")
	}

	tmp, ok := data["supu"]
	if !ok {
		t.Errorf("The proxy returned an unexpected result: %v\n", data)
	}
	supuValue, err := tmp.(json.Number).Int64()
	if err != nil || supuValue != 42 {
		t.Errorf("The proxy returned an unexpected result: %v\n", supuValue)
	}
	if v, ok := data["tupu"]; !ok || !v.(bool) {
		t.Errorf("The proxy returned an unexpected result: %v\n", data)
	}
	if v, ok := data["foo"]; !ok || v.(string) != "bar" {
		t.Errorf("The proxy returned an unexpected result: %v\n", data)
	}
}

func TestNewHTTPProxy_cancel(t *testing.T) {
	expectedMethod := "GET"
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(300) * time.Millisecond)
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{

	}
	request := proxy.Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, err := streamHttpProxy(&backend)(ctx, &request)
	if err == nil || err.Error() != "context deadline exceeded" {
		t.Errorf("The proxy didn't propagate a timeout error: %s\n", err)
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one\n")
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("We were expecting a response at this point in time!\n")
		return
	default:
	}
}

func TestNewHTTPProxy_badResponseBody(t *testing.T) {
	expectedMethod := "GET"
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "supu")
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{}

	request := proxy.Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, _ := streamHttpProxy(&backend)(ctx, &request)

	if response == nil {
		t.Errorf("We were expecting a response but we got no one\n")
	}
	select {
	case <-mustEnd:
		t.Errorf("Error: expected response")
	default:
	}
}

func TestNewHTTPProxy_badStatusCode(t *testing.T) {
	expectedMethod := "GET"
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "booom", 500)
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{
	}
	request := proxy.Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, err := streamHttpProxy(&backend)(ctx, &request)
	if err == nil || err != proxy.ErrInvalidStatusCode {
		t.Errorf("The proxy didn't propagate the backend error: %s\n", err)
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one\n")
	}
	select {
	case <-mustEnd:
		t.Errorf("Error: expected response")
	default:
	}
}

func TestNewHTTPProxy_decodingError(t *testing.T) {
	expectedMethod := "GET"
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"supu": 42}`)
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{

	}
	request := proxy.Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, _ := streamHttpProxy(&backend)(ctx, &request)

	if response == nil {
		t.Errorf("We weren expecting a response but we got no one: \n")
	}
	select {
	case <-mustEnd:
		t.Errorf("Error: expected response")
	default:
	}
}

func TestNewHTTPProxy_badMethod(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("The handler shouldn't be called")
		return
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{
	}
	request := proxy.Request{
		Method: "\n",
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	_, err := streamHttpProxy(&backend)(ctx, &request)
	if err == nil {
		t.Error("The proxy didn't return the expected error")
		return
	}
	if err.Error() != "net/http: invalid method \"\\n\"" {
		t.Errorf("The proxy returned an unexpected error: %s\n", err.Error())
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("Error: expected response")
	default:
	}
}

func TestNewHTTPProxy_requestKo(t *testing.T) {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("The handler shouldn't be called")
		return
	}))
	defer backendServer.Close()

	rpURL, _ := url.Parse(backendServer.URL)
	backend := config.Backend{
	}
	request := proxy.Request{
		Method: "GET",
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()

	expectedError := fmt.Errorf("MAYDAY, MAYDAY")
	_, err := NewHTTPStreamProxyWithHTTPExecutor(&backend, func(_ context.Context, _ *http.Request) (*http.Response, error) {
		return nil, expectedError
	})(ctx, &request)
	if err == nil {
		t.Error("The proxy didn't return the expected error")
		return
	}
	if err != expectedError {
		t.Errorf("The proxy returned an unexpected error: %s\n", err.Error())
		return
	}
	select {
	case <-mustEnd:
		t.Errorf("Error: expected response")
	default:
	}
}

func TestNewRequestBuilderMiddleware_ok(t *testing.T) {
	expected := errors.New("error to be propagated")
	expectedMethod := "GET"
	expectedPath := "/supu"
	assertion := func(ctx context.Context, request *proxy.Request) (*proxy.Response, error) {
		if request.Method != expectedMethod {
			err := fmt.Errorf("Wrong request method. Want: %s. Have: %s", expectedMethod, request.Method)
			t.Errorf(err.Error())
			return nil, err
		}
		if request.Path != expectedPath {
			err := fmt.Errorf("Wrong request path. Want: %s. Have: %s", expectedPath, request.Path)
			t.Errorf(err.Error())
			return nil, err
		}
		return nil, expected
	}
	sampleBackend := config.Backend{
		URLPattern: expectedPath,
		Method:     expectedMethod,
	}
	mw := proxy.NewRequestBuilderMiddleware(&sampleBackend)
	response, err := mw(assertion)(context.Background(), &proxy.Request{})
	if err != expected {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one:\n")
	}
}

func TestNewRequestBuilderMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	sampleBackend := config.Backend{}
	mw := proxy.NewRequestBuilderMiddleware(&sampleBackend)
	mw(explosiveProxy(t), explosiveProxy(t))
}

func explosiveProxy(t *testing.T) proxy.Proxy {
	return func(ctx context.Context, _ *proxy.Request) (*proxy.Response, error) {
		t.Error("This proxy shouldn't been executed!")
		return &proxy.Response{}, nil
	}
}
