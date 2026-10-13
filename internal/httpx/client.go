package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client sends requests to one provider API.
type Client struct {
	HTTPClient *http.Client
	// BaseURL is prefixed to every path, without a trailing slash.
	BaseURL string
	// Authorize sets credentials on each request.
	Authorize func(*http.Request)
	UserAgent string
	// NotFound classifies answers that mean "no such resource" beyond a
	// plain 404 - some APIs answer a missing id with 400 and a message.
	NotFound func(status int, message string) bool
	// MaxRetries is how many times a GET is retried after a 429, a 5xx
	// or a network error. Writes are never retried.
	MaxRetries int

	// sleep waits between retries; replaced in tests.
	sleep func(context.Context, time.Duration) error
}

// Do sends one call and decodes a 2xx JSON body into out (when out is
// non-nil and the body isn't empty).
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body Body, out any) error {
	target := c.BaseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	attempts := 1
	if method == http.MethodGet && c.MaxRetries > 0 {
		attempts += c.MaxRetries
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			if err := c.wait(ctx, backoff(attempt, lastErr)); err != nil {
				return lastErr
			}
		}
		respBody, status, header, err := c.send(ctx, method, target, body)
		if err != nil {
			lastErr = fmt.Errorf("%s %s failed: %w", method, path, err)
			if ctx.Err() != nil {
				return lastErr
			}
			continue
		}
		if status >= 200 && status < 300 {
			if out != nil && len(bytes.TrimSpace(respBody)) > 0 {
				if err := json.Unmarshal(respBody, out); err != nil {
					return fmt.Errorf("failed to parse response from %s %s: %w", method, path, err)
				}
			}
			return nil
		}
		apiErr := &APIError{
			Method:     method,
			Path:       path,
			StatusCode: status,
			Message:    errorMessage(respBody),
			Body:       string(respBody),
		}
		if status == http.StatusTooManyRequests {
			apiErr.RetryAfter = retryAfter(header.Get("Retry-After"))
		}
		apiErr.notFound = status == http.StatusNotFound || (c.NotFound != nil && c.NotFound(status, apiErr.Message))
		lastErr = apiErr
		// A rate limit asking for a longer wait than this client backs off
		// is left to the caller, which can schedule the retry instead of
		// blocking on it.
		if !retryable(status) || apiErr.RetryAfter > maxBackoff {
			return apiErr
		}
	}
	return lastErr
}

func (c *Client) send(ctx context.Context, method, target string, body Body) ([]byte, int, http.Header, error) {
	var reader io.Reader
	contentType := ""
	if body != nil {
		var err error
		reader, contentType, err = body.encode()
		if err != nil {
			return nil, 0, nil, err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, target, reader)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to build request: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("Accept", "application/json")
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if c.Authorize != nil {
		c.Authorize(req)
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return respBody, resp.StatusCode, resp.Header, nil
}
