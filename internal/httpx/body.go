package httpx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// Body is a request body: Form or JSON.
type Body interface {
	encode() (io.Reader, string, error)
}

type formBody url.Values

func (b formBody) encode() (io.Reader, string, error) {
	return strings.NewReader(url.Values(b).Encode()), "application/x-www-form-urlencoded", nil
}

// Form sends v as application/x-www-form-urlencoded.
func Form(v url.Values) Body { return formBody(v) }

type jsonBody struct{ v any }

func (b jsonBody) encode() (io.Reader, string, error) {
	encoded, err := json.Marshal(b.v)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encode request body: %w", err)
	}
	return bytes.NewReader(encoded), "application/json", nil
}

// JSON sends v encoded as JSON.
func JSON(v any) Body { return jsonBody{v} }
