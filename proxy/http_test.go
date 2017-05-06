package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/encoding"
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
		Decoder: encoding.JSONDecoder,
	}
	request := Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	result, err := HTTPProxyFactory(http.DefaultClient)(&backend)(context.Background(), &request)
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

	tmp, ok := result.Data["supu"]
	if !ok {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
	supuValue, err := tmp.(json.Number).Int64()
	if err != nil || supuValue != 42 {
		t.Errorf("The proxy returned an unexpected result: %v\n", supuValue)
	}
	if v, ok := result.Data["tupu"]; !ok || !v.(bool) {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
	}
	if v, ok := result.Data["foo"]; !ok || v.(string) != "bar" {
		t.Errorf("The proxy returned an unexpected result: %v\n", result)
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
		Decoder: encoding.JSONDecoder,
	}
	request := Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, err := httpProxy(&backend)(ctx, &request)
	if err == nil || err.Error() != "context deadline exceeded" {
		t.Errorf("The proxy didn't propagate a timeout error: %s\n", err)
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one: %v\n", response)
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
	backend := config.Backend{
		Decoder: encoding.JSONDecoder,
	}
	request := Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, err := httpProxy(&backend)(ctx, &request)
	if err == nil || err.Error() != "invalid character 's' looking for beginning of value" {
		t.Errorf("The proxy didn't propagate the backend error: %s\n", err)
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one: %v\n", response)
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
		Decoder: encoding.JSONDecoder,
	}
	request := Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, err := httpProxy(&backend)(ctx, &request)
	if err == nil || err != ErrInvalidStatusCode {
		t.Errorf("The proxy didn't propagate the backend error: %s\n", err)
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one: %v\n", response)
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
		Decoder: func(_ io.Reader, _ *map[string]interface{}) error {
			return errors.New("booom")
		},
	}
	request := Request{
		Method: expectedMethod,
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	response, err := httpProxy(&backend)(ctx, &request)
	if err == nil || err.Error() != "booom" {
		t.Errorf("The proxy returned an unexpected error: %s\n", err.Error())
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one: %v\n", response)
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
		Decoder: func(_ io.Reader, _ *map[string]interface{}) error {
			t.Error("The decoder shouldn't be called")
			return nil
		},
	}
	request := Request{
		Method: "\n",
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()
	_, err := httpProxy(&backend)(ctx, &request)
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
		Decoder: func(_ io.Reader, _ *map[string]interface{}) error {
			t.Error("The decoder shouldn't be called")
			return nil
		},
	}
	request := Request{
		Method: "GET",
		Path:   "/",
		URL:    rpURL,
		Body:   newDummyReadCloser(""),
	}
	mustEnd := time.After(time.Duration(150) * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Millisecond)
	defer cancel()

	expectedError := fmt.Errorf("MAYDAY, MAYDAY")
	_, err := NewHTTPProxyWithHTTPExecutor(&backend, func(_ context.Context, _ *http.Request) (*http.Response, error) {
		return nil, expectedError
	}, backend.Decoder)(ctx, &request)
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
	assertion := func(ctx context.Context, request *Request) (*Response, error) {
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
	mw := NewRequestBuilderMiddleware(&sampleBackend)
	response, err := mw(assertion)(context.Background(), &Request{})
	if err != expected {
		t.Errorf("The middleware propagated an unexpected error: %s\n", err.Error())
	}
	if response != nil {
		t.Errorf("We weren't expecting a response but we got one: %v\n", response)
	}
}

func TestNewRequestBuilderMiddleware_multipleNext(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic\n")
		}
	}()
	sampleBackend := config.Backend{}
	mw := NewRequestBuilderMiddleware(&sampleBackend)
	mw(explosiveProxy(t), explosiveProxy(t))
}
