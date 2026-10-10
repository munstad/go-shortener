package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	store, err := NewStore("")
	if err != nil {
		t.Fatal(err)
	}
	return NewMux(store)
}

func create(t *testing.T, h http.Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("POST", "/api/links", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCreateRedirectStats(t *testing.T) {
	h := newTestServer(t)

	rec := create(t, h, `{"url":"https://go.dev"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d", rec.Code)
	}
	var resp createResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	for i := 0; i < 2; i++ {
		rec = httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/"+resp.Code, nil))
		if rec.Code != http.StatusFound || rec.Header().Get("Location") != "https://go.dev" {
			t.Fatalf("redirect: got %d %q", rec.Code, rec.Header().Get("Location"))
		}
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/links/"+resp.Code, nil))
	var l Link
	_ = json.Unmarshal(rec.Body.Bytes(), &l)
	if l.Clicks != 2 {
		t.Fatalf("clicks: got %d, want 2", l.Clicks)
	}
}

func TestInvalidURL(t *testing.T) {
	h := newTestServer(t)
	for _, body := range []string{`{"url":"not a url"}`, `{"url":"ftp://x.com"}`, `{}`, `oops`} {
		if rec := create(t, h, body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: got %d, want 400", body, rec.Code)
		}
	}
}

func TestNotFound(t *testing.T) {
	h := newTestServer(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", rec.Code)
	}
}
