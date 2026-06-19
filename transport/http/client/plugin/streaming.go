// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/luraproject/lura/v2/logging"
)

// executePluginHandler runs an HTTP client plugin handler and exposes its output as an
// *http.Response whose Body streams as the handler writes it.
//
// The handler runs in a goroutine writing into an io.Pipe; the call returns as soon as the
// status line and headers are committed (the first WriteHeader/Write/Flush, or the handler
// returning), with the pipe's read end as the body.
//
// The pipe is unbuffered, so a handler that keeps writing parks in Write until the body is
// read or closed; it unwinds once ctx is cancelled (a watcher then closes the read end) or
// the caller reads or closes the body. A caller that drops the response without reading it
// must therefore cancel ctx or close the body to release the goroutine.
func executePluginHandler(ctx context.Context, logger logging.Logger, logPrefix string, handler http.Handler, req *http.Request) *http.Response {
	pr, pw := io.Pipe()
	w := newStreamingResponseWriter(pw)

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error(logPrefix, "client plugin handler panicked:", rec)
				pw.CloseWithError(fmt.Errorf("client plugin handler panicked: %v", rec))
				// Surface the failure as a 5xx if the handler had not committed a
				// status yet; commit is a no-op once the headers are sent.
				w.commit(http.StatusBadGateway)
			} else {
				pw.Close()
				// A handler that returned without writing still commits an implicit
				// 200 so the caller never waits on headers that never come.
				w.commit(http.StatusOK)
			}
		}()
		handler.ServeHTTP(w, req.WithContext(ctx))
	}()

	// Closing the read end unblocks a handler parked on Write with ErrClosedPipe, so the
	// goroutine unwinds on cancellation regardless of whether a consumer reads the body.
	// Gated on done so the watcher never outlives the handler.
	go func() {
		select {
		case <-ctx.Done():
			pr.CloseWithError(ctx.Err())
		case <-done:
		}
	}()

	// Block only until the status line and headers are known; the body keeps streaming
	// through the pipe after we return.
	select {
	case <-w.committed:
		return newStreamingResponse(w.statusCode, w.committedHeader, pr, req)
	case <-ctx.Done():
		// The context was cancelled before the handler committed any headers; the watcher
		// has closed the body so reads yield ctx.Err(). Synthesise a response instead of
		// blocking on a handler that may never write.
		return newStreamingResponse(http.StatusGatewayTimeout, make(http.Header), pr, req)
	}
}

// newStreamingResponse builds the *http.Response returned by executePluginHandler. The body
// length is unknown up front (the handler streams into a pipe), hence ContentLength -1.
func newStreamingResponse(statusCode int, header http.Header, body io.ReadCloser, req *http.Request) *http.Response {
	return &http.Response{
		StatusCode:    statusCode,
		Status:        fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        header,
		Body:          body,
		ContentLength: -1,
		Request:       req,
	}
}

// streamingResponseWriter is an http.ResponseWriter that pipes everything written to it
// into an *http.Response body. It records the status code and a snapshot of the headers
// the first time the handler commits them (WriteHeader, or the first Write).
type streamingResponseWriter struct {
	header          http.Header
	pw              *io.PipeWriter
	statusCode      int
	committedHeader http.Header
	once            sync.Once
	committed       chan struct{}
}

func newStreamingResponseWriter(pw *io.PipeWriter) *streamingResponseWriter {
	return &streamingResponseWriter{
		header:     make(http.Header),
		pw:         pw,
		statusCode: http.StatusOK,
		committed:  make(chan struct{}),
	}
}

func (w *streamingResponseWriter) Header() http.Header {
	return w.header
}

func (w *streamingResponseWriter) WriteHeader(statusCode int) {
	w.commit(statusCode)
}

func (w *streamingResponseWriter) Write(p []byte) (int, error) {
	// An implicit 200 OK is committed on the first write, like net/http.
	w.commit(http.StatusOK)
	return w.pw.Write(p)
}

// Flush commits the status line and headers if the handler has not done so yet, mirroring
// net/http's implicit WriteHeader(200) on Flush, so a handler that flushes before its first
// write still unblocks the caller. The pipe is unbuffered, so body bytes are already
// delivered as they are written.
func (w *streamingResponseWriter) Flush() {
	w.commit(http.StatusOK)
}

// commit freezes the status code and headers exactly once and signals the waiting caller.
func (w *streamingResponseWriter) commit(statusCode int) {
	w.once.Do(func() {
		w.statusCode = statusCode
		w.committedHeader = w.header.Clone()
		close(w.committed)
	})
}

var _ http.ResponseWriter = (*streamingResponseWriter)(nil)
var _ http.Flusher = (*streamingResponseWriter)(nil)
