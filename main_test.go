package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	t.Setenv("PROJECTS_DB_PATH", filepath.Join(t.TempDir(), "projects.json"))
	r := httptest.NewRecorder()
	newHandler().ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/health", nil))
	if r.Code != http.StatusOK || r.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected response: %d %s", r.Code, r.Body.String())
	}
}

func TestProjectsRequireTenant(t *testing.T) {
	t.Setenv("PROJECTS_DB_PATH", filepath.Join(t.TempDir(), "projects.json"))
	r := httptest.NewRecorder()
	newHandler().ServeHTTP(r, httptest.NewRequest(http.MethodGet, "/projects", nil))
	if r.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", r.Code)
	}
}

func TestProjectsAreIsolatedByTenantAndPersistAcrossRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "projects.json")
	t.Setenv("PROJECTS_DB_PATH", path)
	h := newHandler()
	create := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewBufferString(`{"name":"Alpha"}`))
	create.Header.Set("X-Tenant-ID", "tenant-a")
	created := httptest.NewRecorder()
	h.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("create returned %d: %s", created.Code, created.Body.String())
	}

	h = newHandler()
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

func TestProjectsRejectInvalidTenantAndMalformedInput(t *testing.T) {
	t.Setenv("PROJECTS_DB_PATH", filepath.Join(t.TempDir(), "projects.json"))
	h := newHandler()
	for _, tc := range []struct {
		name, tenant, body string
	}{
		{name: "blank tenant", tenant: " ", body: `{"name":"Alpha"}`},
		{name: "malformed JSON", tenant: "tenant-a", body: `{"name":`},
		{name: "unknown field", tenant: "tenant-a", body: `{"name":"Alpha","tenant":"tenant-b"}`},
		{name: "trailing JSON", tenant: "tenant-a", body: `{"name":"Alpha"}{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(tc.body))
			req.Header.Set("X-Tenant-ID", tc.tenant)
			got := httptest.NewRecorder()
			h.ServeHTTP(got, req)
			if got.Code != http.StatusBadRequest {
				t.Fatalf("got %d, want 400: %s", got.Code, got.Body.String())
			}
		})
	}
}

func TestStorageFailuresDoNotExposeDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-a-directory", "projects.json")
	if err := os.WriteFile(filepath.Dir(path), []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PROJECTS_DB_PATH", path)
	h := newHandler()
	req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(`{"name":"Alpha"}`))
	req.Header.Set("X-Tenant-ID", "tenant-a")
	got := httptest.NewRecorder()
	h.ServeHTTP(got, req)
	if got.Code != http.StatusInternalServerError || strings.Contains(got.Body.String(), "not-a-directory") {
		t.Fatalf("storage failure leaked details or returned wrong status: %d %s", got.Code, got.Body.String())
	}

}
