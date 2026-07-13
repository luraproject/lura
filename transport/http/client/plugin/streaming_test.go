// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/luraproject/lura/v2/logging"
)

func TestExecutePluginHandler_StreamsIncrementally(t *testing.T) {
	release := make(chan struct{})
	handlerDone := make(chan struct{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		defer close(handlerDone)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, "data: first\n\n"); err != nil {
			t.Errorf("write first: %v", err)
			return
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// Block: the handler has NOT returned. If the executor buffered the response, the
		// caller below could not read "first" yet.
		<-release
		if _, err := io.WriteString(w, "data: second\n\n"); err != nil {
			t.Errorf("write second: %v", err)
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/sse", http.NoBody)
	resp := executePluginHandler(context.Background(), logging.NoOp, "", handler, req)

	if resp.StatusCode != http.StatusOK {
		close(release)
		t.Errorf("expected status 200, got %d", resp.StatusCode)
		return
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		close(release)
		t.Errorf("expected text/event-stream, got %q", ct)
		return
	}

	// Read the first event while the handler is still blocked on <-release.
	first := make([]byte, len("data: first\n\n"))
	if _, err := io.ReadFull(resp.Body, first); err != nil {
		close(release)
		t.Errorf("reading first event: %v", err)
		return
	}
	if string(first) != "data: first\n\n" {
		close(release)
		t.Errorf("expected first event, got %q", string(first))
		return
	}

	select {
	case <-handlerDone:
		close(release)
		t.Error("handler returned before we read the first event, response was buffered")
		return
	default:
	}

	// Let the handler finish and emit the second event.
	close(release)
	second := make([]byte, len("data: second\n\n"))
	if _, err := io.ReadFull(resp.Body, second); err != nil {
		t.Errorf("reading second event: %v", err)
		return
	}
	if string(second) != "data: second\n\n" {
		t.Errorf("expected second event, got %q", string(second))
		return
	}

	if rest, err := io.ReadAll(resp.Body); err != nil || len(rest) != 0 {
		t.Errorf("expected clean EOF after stream, got %q err=%v", string(rest), err)
		return
	}
	if err := resp.Body.Close(); err != nil {
		t.Errorf("close body: %v", err)
	}
}

func TestExecutePluginHandler_NoBody(t *testing.T) {
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)
	resp := executePluginHandler(context.Background(), logging.NoOp, "", handler, req)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected implicit 200, got %d", resp.StatusCode)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read body: %v", err)
		return
	}
	if len(body) != 0 {
		t.Errorf("expected empty body, got %q", string(body))
		return
	}
	_ = resp.Body.Close()
}

func TestExecutePluginHandler_NonOKStatus(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Err", "1")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, "boom"); err != nil {
			t.Errorf("write body: %v", err)
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)
	resp := executePluginHandler(context.Background(), logging.NoOp, "", handler, req)

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", resp.StatusCode)
		return
	}
	if resp.Status != "500 Internal Server Error" {
		t.Errorf("expected status line %q, got %q", "500 Internal Server Error", resp.Status)
		return
	}
	if got := resp.Header.Get("X-Err"); got != "1" {
		t.Errorf("expected X-Err header, got %q", got)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read body: %v", err)
		return
	}
	if string(body) != "boom" {
		t.Errorf("expected body %q, got %q", "boom", string(body))
		return
	}
	_ = resp.Body.Close()
}

func TestExecutePluginHandler_HeaderSnapshotFrozenAtCommit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Before", "yes")
		w.WriteHeader(http.StatusCreated)
		// Everything below happens after the first commit and must be ignored.
		w.Header().Set("X-After", "nope")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, "body"); err != nil {
			t.Errorf("write body: %v", err)
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)
	resp := executePluginHandler(context.Background(), logging.NoOp, "", handler, req)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected first committed status 201, got %d", resp.StatusCode)
		return
	}
	if got := resp.Header.Get("X-Before"); got != "yes" {
		t.Errorf("expected X-Before in the frozen snapshot, got %q", got)
		return
	}
	if got := resp.Header.Get("X-After"); got != "" {
		t.Errorf("expected X-After absent from the frozen snapshot, got %q", got)
		return
	}
	if _, err := io.ReadAll(resp.Body); err != nil {
		t.Errorf("read body: %v", err)
		return
	}
	_ = resp.Body.Close()
}

func TestExecutePluginHandler_FlushCommitsHeaders(t *testing.T) {
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Stream", "ready")
		if f, ok := w.(http.Flusher); ok {
			f.Flush() // commit status+headers without WriteHeader/Write
		}
		<-release // block before writing any body
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)

	done := make(chan *http.Response, 1)
	go func() { done <- executePluginHandler(context.Background(), logging.NoOp, "", handler, req) }()

	select {
	case resp := <-done:
		if resp.StatusCode != http.StatusOK {
			close(release)
			t.Errorf("expected implicit 200 after flush, got %d", resp.StatusCode)
			return
		}
		if got := resp.Header.Get("X-Stream"); got != "ready" {
			close(release)
			t.Errorf("expected header committed by Flush, got %q", got)
			return
		}
		close(release)
		_ = resp.Body.Close()
	case <-time.After(2 * time.Second):
		close(release)
		t.Error("executePluginHandler did not return after Flush; Flush must commit headers")
	}
}

func TestExecutePluginHandler_ContextCancellationUnblocksWriter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handlerReturned := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		defer close(handlerReturned)
		w.WriteHeader(http.StatusOK)
		if _, err := io.WriteString(w, "chunk"); err != nil {
			t.Errorf("write first chunk: %v", err)
			return
		}
		// This write blocks: nobody reads it. It must fail once the context is cancelled
		// and the watcher closes the pipe, rather than block forever.
		if _, err := io.WriteString(w, "blocked"); err == nil {
			t.Errorf("expected write to fail after context cancellation")
		}
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)
	resp := executePluginHandler(ctx, logging.NoOp, "", handler, req)

	first := make([]byte, len("chunk"))
	if _, err := io.ReadFull(resp.Body, first); err != nil {
		t.Errorf("reading first chunk: %v", err)
		return
	}

	cancel()

	select {
	case <-handlerReturned:
	case <-time.After(2 * time.Second):
		t.Error("handler goroutine did not unblock after context cancellation")
	}
	_ = resp.Body.Close()
}

func TestExecutePluginHandler_ContextCancelledBeforeCommit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := make(chan struct{})
	unblock := make(chan struct{})
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		close(started)
		// Block without ever committing a status, so only ctx cancellation can return.
		<-unblock
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)

	done := make(chan *http.Response, 1)
	go func() { done <- executePluginHandler(ctx, logging.NoOp, "", handler, req) }()

	<-started
	cancel() // cancel before the handler commits any status

	select {
	case resp := <-done:
		if resp.StatusCode != http.StatusGatewayTimeout {
			close(unblock)
			t.Errorf("expected 504 when ctx is cancelled before commit, got %d", resp.StatusCode)
			return
		}
		close(unblock) // let the handler goroutine unwind
		_ = resp.Body.Close()
	case <-time.After(2 * time.Second):
		close(unblock)
		t.Error("executePluginHandler did not return after ctx cancellation before commit")
	}
}

func TestExecutePluginHandler_PanicBeforeWrite(t *testing.T) {
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("boom")
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)
	resp := executePluginHandler(context.Background(), logging.NoOp, "", handler, req)

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502 on panic before any write, got %d", resp.StatusCode)
		return
	}
	_, err := io.ReadAll(resp.Body)
	if err == nil || !strings.Contains(err.Error(), "client plugin handler panicked") {
		t.Errorf("expected panic surfaced as body error, got %v", err)
		return
	}
	_ = resp.Body.Close()
}

func TestExecutePluginHandler_PanicAfterPartialWrite(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		if _, err := io.WriteString(w, "partial"); err != nil {
			t.Errorf("write partial: %v", err)
			return
		}
		panic("boom")
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/", http.NoBody)
	resp := executePluginHandler(context.Background(), logging.NoOp, "", handler, req)

	if resp.StatusCode != http.StatusTeapot {
		t.Errorf("expected committed status 418 preserved across panic, got %d", resp.StatusCode)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if !strings.HasPrefix(string(body), "partial") {
		t.Errorf("expected partial body before panic, got %q", string(body))
		return
	}
	if err == nil || !strings.Contains(err.Error(), "client plugin handler panicked") {
		t.Errorf("expected panic surfaced as body error after partial body, got %v", err)
		return
	}
	_ = resp.Body.Close()
}
