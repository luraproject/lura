// SPDX-License-Identifier: Apache-2.0

/*
Package gin provides some basic implementations for building routers based on gin-gonic/gin
*/
package gin

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

var _ http.ResponseWriter = (*safeCast)(nil)
var _ http.Flusher = (*safeCast)(nil)
var _ http.Hijacker = (*safeCast)(nil)
var _ http.CloseNotifier = (*safeCast)(nil)

var _ http.Handler = (*safeCaster)(nil)

// safeCast provides fallback implementation for interfaces that are not
// checked internally by gin, and that can cause a panic:
// - Flusher
// - Hijacker
// - Notifier
type safeCast struct {
	w http.ResponseWriter
}

func (s *safeCast) Header() http.Header {
	return s.w.Header()
}

func (s *safeCast) Write(b []byte) (int, error) {
	return s.w.Write(b)
}

func (s *safeCast) WriteHeader(statusCode int) {
	s.w.WriteHeader(statusCode)
}

func (s *safeCast) Flush() {
	if f, ok := s.w.(http.Flusher); ok {
		f.Flush()
	}
}

func (s *safeCast) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := s.w.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, errors.New("not supported")
}

func (s *safeCast) CloseNotify() <-chan bool {
	if h, ok := s.w.(http.CloseNotifier); ok {
		return h.CloseNotify()
	}
	return make(<-chan bool, 1)
}

type safeCaster struct {
	h http.Handler
}

func (s *safeCaster) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(&safeCast{w}, r)
}
