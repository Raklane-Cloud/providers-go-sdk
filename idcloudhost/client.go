package idcloudhost

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"
)

// DefaultBaseURL is the production API, version 1.
const DefaultBaseURL = "https://api.idcloudhost.com/v1"

// Client calls the IDCloudHost API. Set the exported fields before the
// first call; a Client is safe for concurrent use after that.
type Client struct {
	// APIKey is an API token from the IDCloudHost console, sent in the
	// apikey header.
	APIKey string
	// BaseURL defaults to DefaultBaseURL.
	BaseURL string
	// Location is the slug of the location location-specific calls go
	// to (see ListLocations), e.g. "jkt01". Empty uses the account's
	// default location.
	Location string
	// BillingAccountID is charged for created resources when a request
	// doesn't name one. Required unless the API token is restricted to a
	// single billing account.
	BillingAccountID int
	// HTTPClient defaults to a client with a 60s timeout.
	HTTPClient *http.Client
	UserAgent  string

	once sync.Once
	http *httpx.Client
}

// New returns a Client for apiKey in location ("" for the account's
// default location).
func New(apiKey, location string) *Client {
	return &Client{APIKey: apiKey, Location: location}
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
			userAgent = "providers-go-sdk/idcloudhost"
		}
		apiKey := c.APIKey
		c.http = &httpx.Client{
			HTTPClient: httpClient,
			BaseURL:    strings.TrimRight(base, "/"),
			UserAgent:  userAgent,
			Authorize:  func(r *http.Request) { r.Header.Set("apikey", apiKey) },
			NotFound:   notFound,
			MaxRetries: 2,
		}
	})
	return c.http
}

// notFound recognises the API's "no such ..." answers, which don't
// always come with a 404.
func notFound(_ int, message string) bool {
	m := strings.ToLower(message)
	return strings.Contains(m, "no such") || strings.Contains(m, "not found") || strings.Contains(m, "does not exist")
}

// located prefixes path with the client's location, for the API's
// location-specific endpoints.
func (c *Client) located(path string) string {
	if c.Location == "" {
		return path
	}
	return "/" + url.PathEscape(c.Location) + path
}

func (c *Client) billingAccount(id int) int {
	if id != 0 {
		return id
	}
	return c.BillingAccountID
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values, body httpx.Body, out any) error {
	return c.api().Do(ctx, method, path, query, body, out)
}
