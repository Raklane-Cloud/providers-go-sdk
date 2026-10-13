package httpx

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxBackoff = 10 * time.Second

func (c *Client) wait(ctx context.Context, d time.Duration) error {
	if c.sleep != nil {
		return c.sleep(ctx, d)
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusInternalServerError:
		return true
	}
	return false
}

// backoff is the wait before retry number attempt: what a 429 asked for,
// otherwise 500ms doubling, capped at maxBackoff.
func backoff(attempt int, lastErr error) time.Duration {
	var apiErr *APIError
	if errors.As(lastErr, &apiErr) && apiErr.RetryAfter > 0 {
		return min(apiErr.RetryAfter, maxBackoff)
	}
	return min(500*time.Millisecond<<(attempt-1), maxBackoff)
}

// retryAfter parses a Retry-After header given in seconds. Anything else
// is 0.
func retryAfter(header string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(header))
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
