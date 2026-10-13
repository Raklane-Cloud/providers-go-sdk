package biznetgio

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// request is what the fake API saw.
type request struct {
	Method string
	Path   string
	Query  url.Values
	JSON   map[string]any
	Token  string
}

// fakeAPI answers each "METHOD /path" with a status and body, and records
// every request.
func fakeAPI(t *testing.T, routes map[string]func() (int, string)) (*Client, *[]request) {
	t.Helper()
	var seen []request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		req := request{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Token: r.Header.Get("x-token")}
		if r.Header.Get("Content-Type") == "application/json" {
			_ = json.Unmarshal(body, &req.JSON)
		}
		seen = append(seen, req)
		route, ok := routes[r.Method+" "+r.URL.Path]
		if !ok {
			t.Errorf("unexpected call %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		status, resp := route()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(srv.Close)
	c := New("secret")
	c.BaseURL = srv.URL
	return c, &seen
}

func ok(body string) func() (int, string) { return func() (int, string) { return http.StatusOK, body } }
