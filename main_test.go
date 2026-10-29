package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	r := httptest.NewRecorder()
	newHandler().ServeHTTP(r, httptest.NewRequest("GET", "/health", nil))
	if r.Code != http.StatusOK || r.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response: %d %s", r.Code, r.Body.String())
	}
}

func TestProjectsRequireTenant(t *testing.T) {
	r := httptest.NewRecorder()
	newHandler().ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/projects", nil))
	if r.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", r.Code)
	}
}

func TestProjectsAreIsolatedByTenant(t *testing.T) {
	h := newHandler()
	create := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"name":"Alpha"}`))
	create.Header.Set("X-Tenant-ID", "tenant-a")
	created := httptest.NewRecorder()
	h.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", created.Code, created.Body.String())
	}

	for tenant, wantCount := range map[string]int{"tenant-a": 1, "tenant-b": 0} {
		req := httptest.NewRequest(http.MethodGet, "/projects", nil)
		req.Header.Set("X-Tenant-ID", tenant)
		got := httptest.NewRecorder()
		h.ServeHTTP(got, req)
		var projects []project
		if err := json.Unmarshal(got.Body.Bytes(), &projects); err != nil {
			t.Fatal(err)
		}
		if got.Code != http.StatusOK || len(projects) != wantCount {
			t.Fatalf("tenant %s got %d projects, want %d", tenant, len(projects), wantCount)
		}
	}
}
