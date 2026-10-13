package idcloudhost

import "github.com/Raklane-Cloud/providers-go-sdk/internal/httpx"

// APIError is a non-2xx answer from the API. NotFound and Throttled
// classify it; IsNotFound works on wrapped errors.
type APIError = httpx.APIError

// IsNotFound reports whether err means the resource doesn't exist - the
// API answers a missing VM with an error message rather than always a
// 404.
func IsNotFound(err error) bool { return httpx.IsNotFound(err) }
