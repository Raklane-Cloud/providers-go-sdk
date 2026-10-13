package httpx

import (
	"testing"
)

func TestErrorMessageShapes(t *testing.T) {
	for body, want := range map[string]string{
		`{"errors":{"Error":"boom"}}`: "boom",
		`{"errors":["a","b"]}`:        "a; b",
		`{"message":"m"}`:             "m",
		`{"error":"e"}`:               "e",
		`not json`:                    "",
	} {
		if got := errorMessage([]byte(body)); got != want {
			t.Errorf("errorMessage(%s) = %q, want %q", body, got, want)
		}
	}
}
