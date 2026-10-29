package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	r := httptest.NewRecorder()
	mux.ServeHTTP(r, httptest.NewRequest("GET", "/health", nil))
	if r.Code != 200 || r.Body.String() != `{"status":"ok"}` {
		t.Fatalf("unexpected response: %d %s", r.Code, r.Body.String())
	}
}
