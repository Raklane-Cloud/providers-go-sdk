package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func noSleep(context.Context, time.Duration) error { return nil }

func TestDoRetriesGetOnTransientErrors(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, MaxRetries: 2, sleep: noSleep}
	var out struct{ OK bool }
	if err := c.Do(context.Background(), http.MethodGet, "/x", nil, nil, &out); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if !out.OK || calls != 3 {
		t.Fatalf("got ok=%v after %d calls, want true after 3", out.OK, calls)
	}
}

func TestDoLeavesLongRateLimitsToTheCaller(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, MaxRetries: 3, sleep: noSleep}
	err := c.Do(context.Background(), http.MethodGet, "/x", nil, nil, nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.RetryAfter != time.Minute || calls != 1 {
		t.Fatalf("got %v after %d calls, want the 429 returned after 1", err, calls)
	}
}

func TestDoNeverRetriesWrites(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "7")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, MaxRetries: 3, sleep: noSleep}
	err := c.Do(context.Background(), http.MethodPost, "/x", nil, Form(url.Values{"a": {"1"}}), nil)
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !apiErr.Throttled() || apiErr.RetryAfter != 7*time.Second {
		t.Fatalf("got %v, want a throttled APIError with RetryAfter 7s", err)
	}
	if calls != 1 {
		t.Fatalf("POST was sent %d times, want 1", calls)
	}
}

func TestDoClassifiesNotFoundAndParsesMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":{"Error":"No such virtual machine exists. abc"}}`))
	}))
	defer srv.Close()

	c := &Client{
		BaseURL:  srv.URL,
		NotFound: func(_ int, m string) bool { return m == "No such virtual machine exists. abc" },
	}
	err := c.Do(context.Background(), http.MethodGet, "/vm", nil, nil, nil)
	if !IsNotFound(err) {
		t.Fatalf("got %v, want not found", err)
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Message != "No such virtual machine exists. abc" {
		t.Fatalf("message = %q", apiErr.Message)
	}
}

func TestDoSendsBodiesAndAuth(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.Header.Get("apikey") != "k" {
			t.Errorf("apikey header = %q", r.Header.Get("apikey"))
		}
		switch r.URL.Path {
		case "/form":
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" || string(body) != "a=1&b=2" {
				t.Errorf("form body %q (%s)", body, r.Header.Get("Content-Type"))
			}
		case "/json":
			if r.Header.Get("Content-Type") != "application/json" || string(body) != `{"a":1}` {
				t.Errorf("json body %q (%s)", body, r.Header.Get("Content-Type"))
			}
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, Authorize: func(r *http.Request) { r.Header.Set("apikey", "k") }}
	if err := c.Do(context.Background(), http.MethodDelete, "/form", nil, Form(url.Values{"a": {"1"}, "b": {"2"}}), nil); err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := c.Do(context.Background(), http.MethodPost, "/json", nil, JSON(map[string]int{"a": 1}), &out); err != nil {
		t.Fatal(err)
	}
}
