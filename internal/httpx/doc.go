// Package httpx is the small HTTP layer every provider client in this
// module shares: building requests with form or JSON bodies, decoding
// JSON answers, turning non-2xx answers into *APIError, and retrying
// idempotent reads that hit a rate limit or a transient server error.
//
// Writes are never retried here: a create whose answer was lost may
// already have happened, and only the caller knows how to look for it.
package httpx
