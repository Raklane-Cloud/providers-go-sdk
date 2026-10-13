package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// APIError is a non-2xx answer from a provider's API.
type APIError struct {
	Method     string
	Path       string
	StatusCode int
	// Message is the provider's own error text when the body had one in
	// a shape Client recognises, otherwise "".
	Message string
	// Body is the raw response body, for diagnosis.
	Body string
	// RetryAfter is how long a 429 answer asked to wait; 0 when it didn't
	// say.
	RetryAfter time.Duration

	notFound bool
}

func (e *APIError) Error() string {
	detail := e.Message
	if detail == "" {
		detail = e.Body
	}
	return fmt.Sprintf("%s %s returned status %d: %s", e.Method, e.Path, e.StatusCode, detail)
}

// NotFound reports whether the resource the call named doesn't exist.
func (e *APIError) NotFound() bool { return e.notFound }

// Throttled reports whether the provider rate-limited the call.
func (e *APIError) Throttled() bool { return e.StatusCode == http.StatusTooManyRequests }

// IsNotFound reports whether err is an *APIError for a resource that
// doesn't exist.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.NotFound()
}

// errorMessage pulls a human-readable message out of the error bodies
// the providers in this module send:
//
//	{"errors": {"Error": "No such virtual machine exists."}}
//	{"errors": ["..."]}
//	{"message": "..."} / {"error": "..."}
//	{"code": 404, "detail": "..."}
func errorMessage(body []byte) string {
	var shapes struct {
		Errors  json.RawMessage `json:"errors"`
		Message string          `json:"message"`
		Error   any             `json:"error"`
		Detail  any             `json:"detail"`
	}
	if json.Unmarshal(body, &shapes) != nil {
		return ""
	}
	if len(shapes.Errors) > 0 {
		var byKey map[string]any
		if json.Unmarshal(shapes.Errors, &byKey) == nil {
			var parts []string
			for _, v := range byKey {
				parts = append(parts, fmt.Sprint(v))
			}
			if len(parts) > 0 {
				return strings.Join(parts, "; ")
			}
		}
		var list []any
		if json.Unmarshal(shapes.Errors, &list) == nil && len(list) > 0 {
			parts := make([]string, 0, len(list))
			for _, v := range list {
				parts = append(parts, fmt.Sprint(v))
			}
			return strings.Join(parts, "; ")
		}
	}
	if shapes.Message != "" {
		return shapes.Message
	}
	if s, ok := shapes.Error.(string); ok {
		return s
	}
	if s, ok := shapes.Detail.(string); ok {
		return s
	}
	return ""
}
