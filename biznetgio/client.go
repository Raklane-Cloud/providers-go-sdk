package biznetgio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// DefaultBaseURL is the production portal API, version 1.
const DefaultBaseURL = "https://api.portal.biznetgio.com/v1"

// Client calls the Biznet GIO portal API. Set the exported fields before
// the first call; a Client is safe for concurrent use after that.
type Client struct {
	// APIKey is an API token from the Biznet GIO portal, sent in the
	// x-token header.
	APIKey string
	// BaseURL defaults to DefaultBaseURL.
	BaseURL string
	// HTTPClient defaults to a client with a 60s timeout.
	HTTPClient *http.Client
	UserAgent  string

	once sync.Once
	http *httpx.Client
}

// New returns a Client for apiKey.
func New(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

func (c *Client) api() *httpx.Client {
	c.once.Do(func() {
		base := c.BaseURL
		if base == "" {
			base = DefaultBaseURL
		}
		httpClient := c.HTTPClient
		if httpClient == nil {
			httpClient = &http.Client{Timeout: 60 * time.Second}
		}
		userAgent := c.UserAgent
		if userAgent == "" {
			userAgent = "providers-go-sdk/biznetgio"
		}
		apiKey := c.APIKey
		c.http = &httpx.Client{
			HTTPClient: httpClient,
			BaseURL:    strings.TrimRight(base, "/"),
			UserAgent:  userAgent,
			Authorize:  func(r *http.Request) { r.Header.Set("x-token", apiKey) },
			NotFound:   notFound,
			MaxRetries: 2,
		}
	})
	return c.http
}

// notFound recognises "not found" answers that don't come with a 404.
func notFound(_ int, message string) bool {
	m := strings.ToLower(message)
	return strings.Contains(m, "not found") || strings.Contains(m, "does not exist")
}

// do sends one call and decodes the answer's data into out. Answers come
// wrapped as {"success": ..., "code": ..., "data": ...}; a few list
// endpoints send the payload bare, which is decoded as is.
func (c *Client) do(ctx context.Context, method, path string, query url.Values, body httpx.Body, out any) error {
	var raw json.RawMessage
	if err := c.api().Do(ctx, method, path, query, body, &raw); err != nil {
		return err
	}
	if out == nil || len(bytes.TrimSpace(raw)) == 0 {
		return nil
	}
	var envelope map[string]json.RawMessage
	if json.Unmarshal(raw, &envelope) == nil {
		if data, ok := envelope["data"]; ok {
			raw = data
		}
	}
	if string(bytes.TrimSpace(raw)) == "null" {
		return nil // "data": null - nothing, e.g. an empty list
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("failed to parse response from %s %s: %w", method, path, err)
	}
	return nil
}
