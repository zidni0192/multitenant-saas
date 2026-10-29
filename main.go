package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	_ = http.ListenAndServe(":8080", newHandler())
}

type projectRepository interface {
	create(tenant, name string) (project, error)
	list(tenant string) []project
}

func newHandler() http.Handler {
	path := strings.TrimSpace(os.Getenv("PROJECTS_DB_PATH"))
	if path == "" {
		path = filepath.Join("data", "projects.json")
	}
	store, err := openProjectStore(path)
	if err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "storage unavailable", http.StatusInternalServerError)
		})
	}
	return newHandlerWithStore(store)
}

func newHandlerWithStore(store projectRepository) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		tenant, valid := tenantID(r.Header.Get("X-Tenant-ID"))
		if !valid {
			http.Error(w, "tenant identifier is required", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(store.list(tenant))
		case http.MethodPost:
			var input struct {
				Name string `json:"name"`
			}
			decoder := json.NewDecoder(r.Body)
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(&input); err != nil || strings.TrimSpace(input.Name) == "" {
				http.Error(w, "valid project name is required", http.StatusBadRequest)
				return
			}
			if err := decoder.Decode(&struct{}{}); err != io.EOF {
				http.Error(w, "valid project name is required", http.StatusBadRequest)
				return
			}
			created, err := store.create(tenant, strings.TrimSpace(input.Name))
			if err != nil {
				http.Error(w, "storage unavailable", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	return mux
}
